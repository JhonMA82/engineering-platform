package materializer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallBinaryReplacesAtomically(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "eng")
	if err := os.WriteFile(dest, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := InstallBinary([]byte("new binary"), dest); err != nil {
		t.Fatalf("InstallBinary: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil || string(got) != "new binary" {
		t.Fatalf("InstallBinary left %q, %v", got, err)
	}
	if err := InstallBinary(nil, dest); err == nil {
		t.Fatal("InstallBinary(empty) should fail and leave dest untouched")
	}
	got, _ = os.ReadFile(dest)
	if string(got) != "new binary" {
		t.Fatalf("failed InstallBinary mutated dest: %q", got)
	}
}

func TestFetchURLRejectsNon200WithoutLoggingToken(t *testing.T) {
	var seenAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("rate limited"))
	}))
	defer srv.Close()
	_, err := FetchURL(context.Background(), srv.URL, "secret-token", "")
	if err == nil || !strings.Contains(err.Error(), "HTTP 403") {
		t.Fatalf("FetchURL should report HTTP 403, got %v", err)
	}
	if strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("FetchURL leaked the token into the error: %v", err)
	}
	if seenAuth != "Bearer secret-token" {
		t.Fatalf("FetchURL did not send the token, got %q", seenAuth)
	}
}

func TestRunSelfUpdateVerifiesBeforeReplacing(t *testing.T) {
	binary := []byte("fake eng binary v1.4.0")
	sum := sha256.Sum256(binary)
	digest := hex.EncodeToString(sum[:])
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "checksums.txt"):
			_, _ = w.Write([]byte(digest + "  eng-linux-amd64\n"))
		case strings.HasSuffix(r.URL.Path, "eng-linux-amd64"):
			_, _ = w.Write(binary)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// --check verifies everything short of the replace.
	dest := filepath.Join(t.TempDir(), "eng")
	check, err := RunSelfUpdate(context.Background(), SelfUpdateOptions{
		Repo: "O/R", Tag: "v1.4.0", Current: "1.3.0",
		GOOS: "linux", GOARCH: "amd64",
		ExePath: dest, CheckOnly: true, DownloadBase: srv.URL,
	})
	if err != nil {
		t.Fatalf("RunSelfUpdate(check): %v", err)
	}
	if check.Applied {
		t.Fatal("check-only run must not apply")
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatal("check-only run must not create the destination")
	}

	// Applied run replaces the destination atomically.
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	applied, err := RunSelfUpdate(context.Background(), SelfUpdateOptions{
		Repo: "O/R", Tag: "v1.4.0", Current: "1.3.0",
		GOOS: "linux", GOARCH: "amd64",
		ExePath: dest, DownloadBase: srv.URL,
	})
	if err != nil {
		t.Fatalf("RunSelfUpdate(apply): %v", err)
	}
	if !applied.Applied {
		t.Fatal("applied run should report Applied")
	}
	got, err := os.ReadFile(dest)
	if err != nil || string(got) != string(binary) {
		t.Fatalf("destination holds %q, %v", got, err)
	}
}

func TestRunSelfUpdateRefusesTamperedBinary(t *testing.T) {
	sum := sha256.Sum256([]byte("real bytes"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "checksums.txt") {
			_, _ = w.Write([]byte(hex.EncodeToString(sum[:]) + "  eng-linux-amd64\n"))
			return
		}
		_, _ = w.Write([]byte("tampered bytes"))
	}))
	defer srv.Close()
	dest := filepath.Join(t.TempDir(), "eng")
	if _, err := RunSelfUpdate(context.Background(), SelfUpdateOptions{
		Repo: "O/R", Tag: "v1.4.0", Current: "1.3.0",
		GOOS: "linux", GOARCH: "amd64",
		ExePath: dest, DownloadBase: srv.URL,
	}); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("tampered binary should fail verification, got %v", err)
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatal("tampered binary must never touch the destination")
	}
}
func TestCheckSelfUpdatePinnedTagIsOffline(t *testing.T) {
	rep, err := CheckSelfUpdate(context.Background(), SelfUpdateOptions{
		Repo:    "O/R",
		Tag:     "v1.4.0",
		Current: "1.3.0",
		GOOS:    "linux",
		GOARCH:  "amd64",
		ExePath: filepath.Join(t.TempDir(), "eng"),
	})
	if err != nil {
		t.Fatalf("CheckSelfUpdate(pinned): %v", err)
	}
	if rep.To != "v1.4.0" || rep.Asset != "eng-linux-amd64" {
		t.Fatalf("unexpected report: %+v", rep)
	}
}

func TestResolveSelfUpdateTargetDefaults(t *testing.T) {
	_, tag, asset, dest, err := ResolveSelfUpdateTarget(SelfUpdateOptions{Tag: "v1.4.0", GOOS: "linux", GOARCH: "amd64", ExePath: "/tmp/eng-test"})
	if err != nil {
		t.Fatalf("ResolveSelfUpdateTarget: %v", err)
	}
	if tag != "v1.4.0" || asset != "eng-linux-amd64" || dest != "/tmp/eng-test" {
		t.Fatalf("unexpected resolution: %q %q %q", tag, asset, dest)
	}
	if _, _, _, _, err := ResolveSelfUpdateTarget(SelfUpdateOptions{Tag: "v1.0.0", GOOS: "plan9", GOARCH: "amd64", ExePath: "/tmp/e"}); err == nil {
		t.Fatal("unsupported platform should fail")
	}
}
