//go:build !race

package bcn

// raceEnabled reports whether the race detector is active. Timing-based
// guard rails (see TestSanitizeGuardRails) are skipped under -race: the
// instrumentation makes every operation ~10x slower and the guard rails
// would fail spuriously.
const raceEnabled = false
