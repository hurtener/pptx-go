package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestTimelineAccent_Cycle verifies the milestone accent cycle is deterministic,
// wraps, and clamps a negative index.
func TestTimelineAccent_Cycle(t *testing.T) {
	if timelineAccent(0) != pptx.ColorAccent {
		t.Errorf("accent 0 = %v, want ColorAccent", timelineAccent(0))
	}
	if timelineAccent(1) != pptx.ColorAccentAlt {
		t.Errorf("accent 1 = %v, want ColorAccentAlt", timelineAccent(1))
	}
	if timelineAccent(5) != timelineAccent(0) {
		t.Errorf("accent should cycle (5 == 0)")
	}
	if timelineAccent(-3) != pptx.ColorAccent {
		t.Errorf("negative accent index should clamp to ColorAccent")
	}
}

// TestTimelinePreferredHeight verifies per-lane height plus the band-label strip.
func TestTimelinePreferredHeight(t *testing.T) {
	oneLane := timelinePreferredHeight(Timeline{Milestones: []Milestone{{Position: 0.5, Label: "m"}}})
	if oneLane != tlLaneMinH {
		t.Errorf("one implicit lane height = %d, want %d", oneLane, tlLaneMinH)
	}
	twoLaneBand := timelinePreferredHeight(Timeline{
		Lanes: []TimelineLane{{Milestones: []Milestone{{Label: "a"}}}, {Milestones: []Milestone{{Label: "b"}}}},
		Bands: []TimelineBand{{From: 0, To: 1, Label: "Now"}},
	})
	if twoLaneBand != 2*tlLaneMinH+tlBandLabelH {
		t.Errorf("two-lane + band-label height = %d, want %d", twoLaneBand, 2*tlLaneMinH+tlBandLabelH)
	}
	// A band with no label adds no strip.
	noLabel := timelinePreferredHeight(Timeline{
		Milestones: []Milestone{{Label: "m"}},
		Bands:      []TimelineBand{{From: 0, To: 1}},
	})
	if noLabel != tlLaneMinH {
		t.Errorf("unlabeled band should add no strip: got %d, want %d", noLabel, tlLaneMinH)
	}
}
