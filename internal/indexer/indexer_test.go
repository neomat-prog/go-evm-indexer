package indexer

import (
	"testing"

	"github.com/neomat-prog/go-evm-indexer/internal/config"
)

func TestNextRange(t *testing.T) {
	tests := []struct {
		name                string
		cursor, safe, chunk uint64
		wantFrom, wantTo    uint64
		wantOK              bool
	}{
		{name: "caught up", cursor: 100, safe: 100, chunk: 10},
		{name: "safe behind cursor", cursor: 100, safe: 90, chunk: 10},
		{name: "partial chunk", cursor: 100, safe: 105, chunk: 10, wantFrom: 101, wantTo: 105, wantOK: true},
		{name: "clamped to chunk", cursor: 100, safe: 500, chunk: 10, wantFrom: 101, wantTo: 110, wantOK: true},
		{name: "single block", cursor: 0, safe: 1, chunk: 10, wantFrom: 1, wantTo: 1, wantOK: true},
		{name: "wide chunk on paid tier", cursor: 100, safe: 5_000, chunk: 2_000, wantFrom: 101, wantTo: 2_100, wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to, ok := nextRange(tt.cursor, tt.safe, tt.chunk)
			if from != tt.wantFrom || to != tt.wantTo || ok != tt.wantOK {
				t.Errorf("nextRange(%d, %d, %d) = (%d, %d, %v), want (%d, %d, %v)",
					tt.cursor, tt.safe, tt.chunk, from, to, ok, tt.wantFrom, tt.wantTo, tt.wantOK)
			}
		})
	}
}

// Walking nextRange to completion must cover every block in (cursor, safe]
// exactly once — a gap is silent data loss, an overlap is wasted work — and no
// window may exceed chunk blocks, or the provider rejects the call outright.
func TestNextRangeTilesWithoutGapOrOverlap(t *testing.T) {
	const (
		start = 1_000
		safe  = 1_450
		chunk = 10
	)

	cursor := uint64(start)
	seen := make(map[uint64]int)

	for {
		from, to, ok := nextRange(cursor, safe, chunk)
		if !ok {
			break
		}
		if from != cursor+1 {
			t.Fatalf("gap: cursor %d, next from %d", cursor, from)
		}
		if span := to - from + 1; span > chunk {
			t.Fatalf("window %d-%d spans %d blocks, provider cap is %d", from, to, span, chunk)
		}
		for b := from; b <= to; b++ {
			seen[b]++
		}
		cursor = to
	}

	if cursor != safe {
		t.Fatalf("stopped at %d, want %d", cursor, safe)
	}
	for b := uint64(start + 1); b <= safe; b++ {
		if seen[b] != 1 {
			t.Fatalf("block %d scanned %d times, want 1", b, seen[b])
		}
	}
}

// The two zero values that are fatal, not just wrong.
func TestNewFillsUnsafeZeroValues(t *testing.T) {
	ix := New(nil, nil, config.Config{})

	if ix.cfg.Chunk != config.DefaultChunk {
		t.Errorf("Chunk = %d, want %d", ix.cfg.Chunk, config.DefaultChunk)
	}
	if ix.cfg.Interval != config.DefaultInterval {
		t.Errorf("Interval = %v, want %v", ix.cfg.Interval, config.DefaultInterval)
	}

	set := New(nil, nil, config.Config{Chunk: 2_000})
	if set.cfg.Chunk != 2_000 {
		t.Errorf("New overwrote an explicit Chunk: got %d", set.cfg.Chunk)
	}
}

func TestSaturatingSub(t *testing.T) {
	tests := []struct {
		a, b, want uint64
	}{
		{100, 12, 88},
		{12, 12, 0},
		{3, 12, 0},
		{0, 0, 0},
	}

	for _, tt := range tests {
		if got := saturatingSub(tt.a, tt.b); got != tt.want {
			t.Errorf("saturatingSub(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
