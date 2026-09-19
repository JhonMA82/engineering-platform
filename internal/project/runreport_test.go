package project

import (
	"testing"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

func TestSummarizeIntentKeepsOutputOnGarbage(t *testing.T) {
	sum := SummarizeIntent([]byte("{nope"), "outdir")
	if sum.Output != "outdir" {
		t.Fatalf("output not kept: %+v", sum)
	}
	if sum.Name != "" {
		t.Fatalf("garbage intent should not name anything: %+v", sum)
	}
}

func TestErrorReportCarriesHint(t *testing.T) {
	intent := SummarizeIntent([]byte(`{"schema_version":1,"name":"demo","problem":"ops","surfaces":[{"kind":"api","scope":"required_now"}]}`), "outdir")
	rep := ErrorReport("start", intent, domain.Catalog("boom"), nil, "")
	if rep.Result.Status != RunError {
		t.Fatalf("status = %q", rep.Result.Status)
	}
	if rep.Result.ErrorClass != "catalog" {
		t.Fatalf("class = %q", rep.Result.ErrorClass)
	}
	if rep.Result.Hint == "" {
		t.Fatal("hint must not be empty")
	}
	if rep.Intent.Name != "demo" || len(rep.Intent.Surfaces) != 1 {
		t.Fatalf("intent not projected: %+v", rep.Intent)
	}
}

func TestDoctorReportOKHasNoHint(t *testing.T) {
	rep := DoctorReport(IntentSummary{Output: "proj"}, nil)
	if rep.Result.Status != RunOK {
		t.Fatalf("status = %q", rep.Result.Status)
	}
	if rep.Result.Hint != "" {
		t.Fatalf("clean doctor must not hint: %+v", rep.Result)
	}
}

func TestErrorClassOfFallback(t *testing.T) {
	class, msg := ErrorClassOf(domain.Validation("bad name"))
	if class != "validation" || msg != "bad name" {
		t.Fatalf("got %q %q", class, msg)
	}
}
