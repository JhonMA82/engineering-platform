package materializer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/selfupdate"
)

// selfUpdateTimeout bounds every self-update HTTP fetch: an updater must
// never hang a terminal indefinitely on a stalled network.
const selfUpdateTimeout = 90 * time.Second

// maxDownloadBytes caps a single self-update fetch (binary or checksums).
// Release binaries are a few tens of MB; 256MB leaves wide headroom while
// refusing absurd payloads before they fill the disk.
const maxDownloadBytes = 256 << 20

// SelfUpdateReport describes one resolved or applied binary update.
type SelfUpdateReport struct {
	// From is the running core line ("dev" for unstamped builds).
	From string
	// To is the target release tag ("vX.Y.Z").
	To string
	// Asset is the platform release file that was (or would be) fetched.
	Asset string
	// Dest is the executable path that was (or would be) replaced.
	Dest string
	// UpdateAvailable is false when the running line is already current
	// (self-update never downgrades).
	UpdateAvailable bool
	// Applied is false for --check runs (nothing was written).
	Applied bool
}

// SelfUpdateOptions tunes one update. Tag "" means "latest release";
// CheckOnly resolves and verifies without replacing the binary. ExePath
// overrides os.Executable (tests); Token authenticates GitHub API/file
// downloads when set (never logged).
type SelfUpdateOptions struct {
	Repo      string
	Tag       string
	Current   string
	CheckOnly bool
	ExePath   string
	Token     string
	GOOS      string
	GOARCH    string
	// DownloadBase overrides ReleaseBaseURL(repo, tag) for hermetic
	// tests (an httptest server). Empty means the real release line.
	DownloadBase string
}

// FetchURL downloads url with a bounded timeout and a size cap. Non-200
// responses fail with the status and a short body tail; the token travels
// only in the Authorization header and is never logged. An empty accept
// defaults to application/octet-stream (release binaries); GitHub API
// calls pass application/vnd.github+json instead.
func FetchURL(ctx context.Context, url, token, accept string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, selfUpdateTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, domain.ExternalCommand(fmt.Sprintf("self-update request: %v", err))
	}
	req.Header.Set("User-Agent", "eng-self-update")
	if strings.TrimSpace(accept) == "" {
		accept = "application/octet-stream"
	}
	req.Header.Set("Accept", accept)
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, domain.ExternalCommand(fmt.Sprintf("self-update download: %v", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		peek, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, domain.ExternalCommand(fmt.Sprintf("self-update download %s: HTTP %d\n%s", url, resp.StatusCode, strings.TrimSpace(string(peek))))
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadBytes+1))
	if err != nil {
		return nil, domain.ExternalCommand(fmt.Sprintf("self-update download: %v", err))
	}
	if int64(len(data)) > maxDownloadBytes {
		return nil, domain.ExternalCommand(fmt.Sprintf("self-update download %s: payload exceeds %d bytes", url, maxDownloadBytes))
	}
	return data, nil
}

// LatestTag resolves the newest release tag via the GitHub API.
func LatestTag(ctx context.Context, repo, token string) (string, error) {
	if err := selfupdate.ValidateRepo(repo); err != nil {
		return "", domain.Validation(err.Error())
	}
	raw, err := FetchURL(ctx, selfupdate.LatestReleaseAPIURL(repo), token, "application/vnd.github+json")
	if err != nil {
		return "", err
	}
	tag, err := selfupdate.ParseLatestTag(raw)
	if err != nil {
		return "", domain.ExternalCommand(fmt.Sprintf("self-update: %v", err))
	}
	return selfupdate.NormalizeTag(tag)
}

// InstallBinary atomically replaces dest with data: write a temp file in
// the same directory (same filesystem, so rename is atomic), chmod 0755,
// then rename over dest. Refusal leaves dest untouched. On Windows the
// rename of a running executable fails; the error names the temp file so
// the caller can print the manual step.
func InstallBinary(data []byte, dest string) error {
	if len(data) == 0 {
		return domain.Filesystem("self-update binary is empty")
	}
	if strings.TrimSpace(dest) == "" {
		return domain.Filesystem("self-update destination is empty")
	}
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, ".eng-update-*")
	if err != nil {
		return domain.Filesystem(fmt.Sprintf("self-update stage: %v", err))
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return domain.Filesystem(fmt.Sprintf("self-update stage: %v", err))
	}
	if err := tmp.Close(); err != nil {
		return domain.Filesystem(fmt.Sprintf("self-update stage: %v", err))
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		return domain.Filesystem(fmt.Sprintf("self-update chmod: %v", err))
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return domain.Filesystem(fmt.Sprintf("self-update replace %q: %v (staged file kept at %q; on Windows close every eng process and move it over eng.exe manually)", dest, err, tmpName))
	}
	return nil
}

// ResolveSelfUpdateTarget validates the options and resolves the tag,
// asset and destination without any network or filesystem mutation beyond
// reading the executable path.
func ResolveSelfUpdateTarget(opts SelfUpdateOptions) (repo, tag, asset, dest string, err error) {
	repo = strings.TrimSpace(opts.Repo)
	if repo == "" {
		repo = selfupdate.DefaultRepo
	}
	if verr := selfupdate.ValidateRepo(repo); verr != nil {
		return "", "", "", "", domain.Validation(verr.Error())
	}
	goos, goarch := opts.GOOS, opts.GOARCH
	if goos == "" {
		goos = runtime.GOOS
	}
	if goarch == "" {
		goarch = runtime.GOARCH
	}
	asset, aerr := selfupdate.AssetName(goos, goarch)
	if aerr != nil {
		return "", "", "", "", domain.Validation(aerr.Error())
	}
	tag = strings.TrimSpace(opts.Tag)
	if tag == "" || tag == "latest" {
		tag = "latest"
	} else if norm, nerr := selfupdate.NormalizeTag(tag); nerr != nil {
		return "", "", "", "", domain.Validation(nerr.Error())
	} else {
		tag = norm
	}
	dest = strings.TrimSpace(opts.ExePath)
	if dest == "" {
		exe, eerr := os.Executable()
		if eerr != nil {
			return "", "", "", "", domain.Filesystem(fmt.Sprintf("self-update executable path: %v", eerr))
		}
		if resolved, rerr := filepath.EvalSymlinks(exe); rerr == nil {
			exe = resolved
		}
		dest = exe
	}
	return repo, tag, asset, dest, nil
}

// CheckSelfUpdate resolves the target tag (fetching "latest" when asked)
// and reports whether it is newer than current. It performs no writes.
func CheckSelfUpdate(ctx context.Context, opts SelfUpdateOptions) (*SelfUpdateReport, error) {
	repo, tag, asset, dest, err := ResolveSelfUpdateTarget(opts)
	if err != nil {
		return nil, err
	}
	if tag == "latest" {
		tag, err = LatestTag(ctx, repo, opts.Token)
		if err != nil {
			return nil, err
		}
	}
	update, err := selfupdate.NeedsUpdate(opts.Current, tag)
	if err != nil {
		return nil, domain.Validation(fmt.Sprintf("self-update version check: %v", err))
	}
	return &SelfUpdateReport{From: opts.Current, To: tag, Asset: asset, Dest: dest, UpdateAvailable: update}, nil
}

// RunSelfUpdate downloads the target binary, verifies its sha256 against
// the release checksums.txt, and atomically replaces the running
// executable — unless CheckOnly, which verifies everything short of the
// replace. The binary is never touched before its checksum verifies.
func RunSelfUpdate(ctx context.Context, opts SelfUpdateOptions) (*SelfUpdateReport, error) {
	report, err := CheckSelfUpdate(ctx, opts)
	if err != nil {
		return nil, err
	}
	repo := strings.TrimSpace(opts.Repo)
	if repo == "" {
		repo = selfupdate.DefaultRepo
	}
	base := strings.TrimSuffix(strings.TrimSpace(opts.DownloadBase), "/")
	if base == "" {
		base = selfupdate.ReleaseBaseURL(repo, report.To)
	}
	sumsRaw, err := FetchURL(ctx, base+"/"+selfupdate.ChecksumsFile, opts.Token, "")
	if err != nil {
		return nil, err
	}
	sums, err := selfupdate.ParseChecksums(string(sumsRaw))
	if err != nil {
		return nil, domain.ExternalCommand(fmt.Sprintf("self-update: %v", err))
	}
	want, err := selfupdate.ExpectedChecksum(sums, report.Asset)
	if err != nil {
		return nil, domain.ExternalCommand(fmt.Sprintf("self-update: %v", err))
	}
	bin, err := FetchURL(ctx, base+"/"+report.Asset, opts.Token, "")
	if err != nil {
		return nil, err
	}
	if err := selfupdate.VerifySHA256(bin, want); err != nil {
		return nil, domain.ExternalCommand(fmt.Sprintf("self-update: %v", err))
	}
	if opts.CheckOnly {
		report.Applied = false
		return report, nil
	}
	if err := InstallBinary(bin, report.Dest); err != nil {
		return nil, err
	}
	report.Applied = true
	return report, nil
}
