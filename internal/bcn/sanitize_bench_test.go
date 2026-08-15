package bcn

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// benchInput builds a ~100KB input from the real BCN norma fixture (all
// content blocks concatenated), representative of a large law.
func benchInput(tb testing.TB) string {
	tb.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "norma_full.json"))
	if err != nil {
		tb.Fatal(err)
	}
	var norma NormaFull
	if err := json.Unmarshal(data, &norma); err != nil {
		tb.Fatal(err)
	}
	var b strings.Builder
	for _, block := range norma.Html {
		b.WriteString(block.T)
	}
	in := b.String()
	for len(in) < 100_000 {
		in += b.String()
	}
	return in
}

// TestSanitizeGuardRails is the custodian of the single-pass optimization:
// it measures the sanitizer on real BCN data and fails if a future change
// makes it slower or allocation-heavier than the validated baseline
// (measured in the design spike: 0.63ms/op, 8 allocs per ~108KB).
func TestSanitizeGuardRails(t *testing.T) {
	if testing.Short() {
		t.Skip("guard rails skipped in -short mode")
	}
	if raceEnabled {
		t.Skip("guard rails skipped under -race: timing is not meaningful with race instrumentation")
	}
	input := benchInput(t)
	res := testing.Benchmark(func(b *testing.B) {
		for b.Loop() {
			_ = normalize(input)
		}
	})

	const maxNsPerOp = 2 * time.Millisecond
	const maxAllocsPerOp = 15

	if res.NsPerOp() > int64(maxNsPerOp) {
		t.Errorf("sanitizer too slow: %v per op (limit %v)", res.NsPerOp(), maxNsPerOp)
	}
	if res.AllocsPerOp() > maxAllocsPerOp {
		t.Errorf("sanitizer too allocation-heavy: %v allocs/op (limit %d)", res.AllocsPerOp(), maxAllocsPerOp)
	}
	t.Logf("sanitize: %v per op, %v allocs/op, %d bytes input", res.NsPerOp(), res.AllocsPerOp(), len(input))
}

// BenchmarkNormalize is the raw benchmark for `go test -bench`.
func BenchmarkNormalize(b *testing.B) {
	input := benchInput(b)
	b.ReportAllocs()
	for b.Loop() {
		_ = normalize(input)
	}
}
