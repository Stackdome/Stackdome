package releasegc

import (
	"sort"
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func summaries(specs ...[2]int) []models.ReleaseSummary {
	out := make([]models.ReleaseSummary, 0, len(specs))
	for i, sp := range specs {
		st := models.ReleaseStateFailed
		if sp[1] == 1 {
			st = models.ReleaseStateReleased
		}
		out = append(out, models.ReleaseSummary{ID: string(rune('a' + i)), Sequence: sp[0], State: st})
	}
	return out
}

func ids(s []string) []string { sort.Strings(s); return s }

func TestReleaseIDsToGC_KeepsConvergedFloorAndCapsAtTen(t *testing.T) {
	var specs [][2]int
	for seq := 1; seq <= 5; seq++ {
		specs = append(specs, [2]int{seq, 1})
	}
	for seq := 6; seq <= 20; seq++ {
		specs = append(specs, [2]int{seq, 0})
	}
	all := summaries(specs...)

	del := releaseIDsToGC(all)

	if len(del) != 10 {
		t.Fatalf("expected 10 deletions, got %d", len(del))
	}
	kept := map[string]bool{}
	for _, s := range all {
		kept[s.ID] = true
	}
	for _, d := range del {
		delete(kept, d)
	}
	if len(kept) != 10 {
		t.Fatalf("expected 10 kept, got %d", len(kept))
	}
}

func TestReleaseIDsToGC_NothingToDeleteUnderCap(t *testing.T) {
	all := summaries([2]int{1, 1}, [2]int{2, 0}, [2]int{3, 1})
	if del := releaseIDsToGC(all); len(del) != 0 {
		t.Fatalf("expected no deletions, got %v", ids(del))
	}
}
