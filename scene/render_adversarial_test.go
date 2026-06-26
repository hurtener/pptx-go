package scene_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// R11.12 adversarial content-fit fixtures (D-092). A reusable torture suite that
// renders every component under hostile content — multi-line headers, over-long
// pill / badge / row labels, oversized stat values, many list items, dark variant +
// dark fills — and asserts the structural invariant that no emitted shape/text box
// ever falls off the slide canvas. Catches the whole class of fixed-size regressions
// (any component that assumes a fixed width/height and draws off-slide) for any
// future component, in one place.

const longText = "A deliberately long stretch of content that wraps across several lines under any reasonable column width"

// ptrF returns a pointer to v (for the optional typed Stat.Number path, R14.13).
func ptrF(v float64) *float64 { return &v }

// adversarialScene builds a deck exercising every component with hostile content,
// in both the light and dark variants.
func adversarialScene() scene.Scene {
	accent := scene.ColorAccent
	dotColor := scene.ColorSuccess
	body := func() []scene.SlideNode {
		return []scene.SlideNode{scene.List{Items: []scene.ListItem{
			{Text: rt(longText)}, {Text: rt(longText)}, {Text: rt(longText)}, {Text: rt(longText)},
		}, Indent: scene.IndentTight}}
	}
	card := func(variantDark bool) scene.Card {
		fill := scene.ColorSurface
		if variantDark {
			fill = scene.ColorAccent
		}
		return scene.Card{
			Eyebrow: "VISION", Header: longText, HeaderPill: "FULLY CUSTOMIZABLE",
			HeaderFill: &accent, StatusDot: &dotColor, Watermark: "01", Fill: fill,
			// A top-bar ribbon (R12.3) reserves a band so the long wrapped header below
			// it must still stay on-canvas and clear the body.
			Ribbon: &scene.Ribbon{Text: "MOST POPULAR · THE MOAT", Position: scene.RibbonTopBar},
			Body:   body(),
		}
	}

	var slides []scene.SceneSlide
	for _, variant := range []scene.Variant{scene.VariantLight, scene.VariantDark} {
		dark := variant == scene.VariantDark
		slides = append(slides,
			// Cards (every size × layout) under a long wrapping header.
			scene.SceneSlide{ID: fmt.Sprintf("cards-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Grid{Columns: 3, Cells: []scene.SlideNode{card(dark), card(dark), card(dark)},
					// Inter-column connectors (R12.4) in the gutters must stay on-canvas.
					Connectors: []scene.GridConnector{
						{Between: [2]int{0, 1}, Kind: scene.ConnectorArrow, Label: "feeds"},
						{Between: [2]int{1, 2}, Kind: scene.ConnectorBiArrow},
					}},
			}},
			// A dense bento with long row labels + an over-wide stat.
			scene.SceneSlide{ID: fmt.Sprintf("bento-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Bento{Columns: 2, WeightedRows: true, Rows: []scene.BentoRow{
					{Label: "Control plane orchestration", Cells: []scene.BentoCell{{Span: 1, Node: scene.Stat{Value: "$4,000,000+", Label: "per month", AutoFit: true}}, {Span: 1, Node: scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}}}},
					{Label: "The core memory layer", Cells: []scene.BentoCell{{Span: 2, Node: scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}}}},
					{Label: "Run", Cells: []scene.BentoCell{{Span: 1, Node: scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}}, {Span: 1, Node: scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}}}},
				}},
			}},
			// TwoColumn with a long join badge label.
			scene.SceneSlide{ID: fmt.Sprintf("cols-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.TwoColumn{Join: scene.JoinBadge, JoinLabel: "One agent for everything",
					Left:  []scene.SlideNode{scene.Heading{Text: rt(longText), Level: 1, AutoFit: true}, scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}},
					Right: []scene.SlideNode{scene.Stat{Value: "$4,000+", Label: "up to 100 agents", AutoFit: true}, scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}}},
			}},
			// TwoColumn with a top-bridge join (R12.8): the bracket + label pill must span
			// both columns and stay on-canvas, never wrapping the label mid-word.
			scene.SceneSlide{ID: fmt.Sprintf("bridge-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.TwoColumn{Join: scene.JoinBadge, JoinLabel: "One agent, purpose-built — two ways to get it", JoinPosition: scene.JoinTopBridge,
					Left:  []scene.SlideNode{scene.Heading{Text: rt(longText), Level: 2, AutoFit: true}, scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}},
					Right: []scene.SlideNode{scene.Heading{Text: rt(longText), Level: 2, AutoFit: true}, scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}}},
			}},
			// A tall stat strip + hero that would overflow without the guards, over a
			// gradient legibility scrim (R14.1): the scrim overlay must not displace
			// content and text must stay on-canvas above it.
			scene.SceneSlide{ID: fmt.Sprintf("strip-%v", variant), Variant: variant,
				Background: scene.Background{Kind: scene.BackgroundColor, Color: scene.ColorAccent,
					Scrim: &scene.Scrim{Color: scene.ColorSurface, Opacity: 40000, Gradient: true}},
				// Footnotes (R14.12) reserve a bottom band; the body must shrink + stay clear.
				Footnotes: []scene.RichText{rt(longText), rt("Source: " + longText)},
				Nodes: []scene.SlideNode{
					scene.Hero{Title: longText, AutoFit: true},
					scene.Grid{Columns: 4, Cells: []scene.SlideNode{
						// A typed-number Stat (R14.13): formats to "$4,000,000+" via NumberFormat,
						// staying on one line under AutoFit (the slide-09 wrap regression fix).
						scene.Stat{Number: ptrF(4000000), Format: &scene.NumberFormat{GroupSep: ",", CurrencySymbol: "$", Suffix: "+"}, Label: "a", AutoFit: true},
						scene.Stat{Value: "99.999%", Label: "b", AutoFit: true},
						scene.Stat{Value: "1,234,567", Label: "c", AutoFit: true}, scene.Stat{Value: "$12.5M", Label: "d", AutoFit: true},
					}},
				}},
			// Buttons (R12.1): a long-label primary (forces width clamp + label fitScale),
			// a ghost outline + leading icon, and a button last in a card body (must stay
			// inside the card padding). Centered so the offset path is exercised too.
			scene.SceneSlide{ID: fmt.Sprintf("buttons-%v", variant), Variant: variant,
				Content: scene.Alignment{Horizontal: scene.HAlignCenter},
				Nodes: []scene.SlideNode{
					scene.Button{Label: longText, Tone: scene.ButtonPrimary, Size: scene.ButtonLG, TrailingIcon: "arrow-right"},
					scene.Button{Label: "Talk to the team", Tone: scene.ButtonGhost, LeadingIcon: "star"},
					scene.Card{Header: "Scale", Body: []scene.SlideNode{
						scene.Stat{Value: "$399", Label: "per month"},
						scene.Button{Label: "Start free", Tone: scene.ButtonAccentAlt, TrailingIcon: "arrow-right"},
					}},
				}},
			// A filled 2-column checklist with long items inside a Fill card (R12.2):
			// glyphs stay on-canvas and the text never collides with the glyph or wraps off.
			scene.SceneSlide{ID: fmt.Sprintf("checklist-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Card{Header: "What you get", Fill: scene.ColorSurface, BodyVAlign: scene.VAlignFill, Body: []scene.SlideNode{
					scene.Checklist{Columns: 2, Fill: true, Items: []scene.ChecklistItem{
						{Text: rt(longText), State: scene.CheckDone},
						{Text: rt(longText), State: scene.CheckDone},
						{Text: rt(longText), State: scene.CheckNo},
						{Text: rt(longText), State: scene.CheckNeutral},
						{Text: rt(longText), State: scene.CheckDone},
					}},
				}},
			}},
			// A labeled wrapping chip row with many long chips (R12.5): every pill must
			// wrap within the width and stay on-canvas.
			scene.SceneSlide{ID: fmt.Sprintf("chips-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.ChipRow{Label: "COMMON BUILDS — every department", Wrap: true, Chips: []scene.ChipSpec{
					{Label: "Finance & accounting"}, {Label: "Human resources"}, {Label: "Sales operations"},
					{Label: "Legal & compliance"}, {Label: "Operations"}, {Label: "Reporting & analytics"},
					{Label: "Custom integrations", Tone: scene.ChipSolid, Color: scene.ColorAccent},
				}},
			}},
			// A full-width banner with a long lead/body + an embedded trailing button
			// (R12.6): the strip, text, and button must stay on-canvas and legible.
			scene.SceneSlide{ID: fmt.Sprintf("banner-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Banner{Lead: rt(longText), Body: rt(longText), Icon: "star", Fill: scene.ColorAccent,
					Trailing: []scene.SlideNode{scene.Button{Label: "Talk to the team", Tone: scene.ButtonNeutral, TrailingIcon: "arrow-right"}}},
			}},
			// A centered attribution lockup (R12.9): caption + icon mark stay on-canvas.
			scene.SceneSlide{ID: fmt.Sprintf("lockup-%v", variant), Variant: variant,
				Content: scene.Alignment{Horizontal: scene.HAlignCenter},
				Nodes: []scene.SlideNode{
					scene.Lockup{Caption: "POWERED BY A VERY LONG PARTNER NAME CLEAR TECHNOLOGIES", Icon: "star", AssetSide: scene.TrailCaption},
				}},
			// A filled icon-rows list with long labels + right-aligned meta in a Fill card
			// (R12.7): icons, labels, and meta must stay on-canvas without colliding.
			scene.SceneSlide{ID: fmt.Sprintf("iconrows-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Card{Header: "Integrations", Fill: scene.ColorSurface, BodyVAlign: scene.VAlignFill, Body: []scene.SlideNode{
					scene.IconRows{Fill: true, Rows: []scene.IconRow{
						{Icon: "star", Label: rt(longText), Meta: rt("Microsoft 365 · Workspace"), Tone: scene.RowPill},
						{Icon: "check", Label: rt(longText), Meta: rt("CRM")},
						{Icon: "dot", Label: rt(longText), Tone: scene.RowPill},
					}},
				}},
			}},
			// A dense roadmap (R14.4): 2 lanes × many milestones + phase bands, with
			// long labels. The axis, markers, and staggered labels must stay on-canvas.
			scene.SceneSlide{ID: fmt.Sprintf("roadmap-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Timeline{
					Bands: []scene.TimelineBand{{From: 0, To: 0.5, Label: "NOW · the present horizon", Fill: scene.ColorAccent}, {From: 0.5, To: 1, Label: "NEXT", Fill: scene.ColorInfo}},
					Lanes: []scene.TimelineLane{
						{Label: "Platform engineering", Milestones: []scene.Milestone{
							{Position: 0, Label: longText, Icon: "star"}, {Position: 0.45, Label: "GA", Detail: longText, AccentIndex: 1}, {Position: 1, Label: "Scale", AccentIndex: 2},
						}},
						{Label: "Go-to-market", Milestones: []scene.Milestone{{Position: 0.3, Label: longText}, {Position: 0.95, Label: "Expand"}}},
					},
				},
			}},
			// Native micro-charts (R14.8): a labeled progress bar, a bar group, and a
			// sparkline (with an upward segment) in a Fill card — all must stay on-canvas.
			scene.SceneSlide{ID: fmt.Sprintf("dataviz-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Card{Header: "KPIs", Fill: scene.ColorSurface, BodyVAlign: scene.VAlignFill, Body: []scene.SlideNode{
					scene.DataMark{Kind: scene.DataMarkBar, Value: 0.92, Label: "92%"},
					scene.DataMark{Kind: scene.DataMarkBars, Values: []float64{0.3, 0.6, 0.9, 0.5, 0.7}},
					scene.DataMark{Kind: scene.DataMarkSparkline, Values: []float64{0.2, 0.8, 0.4, 1.0, 0.6, 0.9}},
					scene.DataMark{Kind: scene.DataMarkDonut, Value: 0.92, Label: "92%"},
					scene.DataMark{Kind: scene.DataMarkGauge, Value: 0.66, Label: "66"},
				}},
			}},
			// A funnel + cycle (R14.11): tapering bands + a ring of stage cards with long
			// labels must stay on-canvas.
			scene.SceneSlide{ID: fmt.Sprintf("funnel-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.TwoColumn{
					Left:  []scene.SlideNode{scene.Funnel{Stages: []scene.FunnelStage{{Label: longText, Value: "100k"}, {Label: "Mid", Value: "10k"}, {Label: longText, Value: "1k"}}}},
					Right: []scene.SlideNode{scene.Cycle{Stages: []scene.CycleStage{{Label: longText, Icon: "star"}, {Label: "B"}, {Label: longText}, {Label: "D"}}}},
				},
			}},
			// A hierarchy/org tree (R14.10): a 3-level tree with long labels; node cards
			// + elbow edges must stay on-canvas.
			scene.SceneSlide{ID: fmt.Sprintf("tree-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Tree{Root: scene.TreeNode{Label: longText, Children: []scene.TreeNode{
					{Label: longText, Icon: "star", Children: []scene.TreeNode{{Label: longText}, {Label: "Apps"}}},
					{Label: "GTM", AccentIndex: 2, Children: []scene.TreeNode{{Label: longText}}},
				}}},
			}},
			// A logo wall (R14.7): a mono-tone grid of logos with a caption; the grid
			// cells + contained logos must stay on-canvas. (No resolver → logos warn +
			// skip; the caption + layout still exercise the composer.)
			scene.SceneSlide{ID: fmt.Sprintf("logowall-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.LogoWall{Caption: longText, Columns: 4, Tone: scene.LogoToneMono, Logos: []scene.LogoEntry{
					{AssetID: "asset://x1"}, {AssetID: "asset://x2"}, {AssetID: "asset://x3"},
					{AssetID: "asset://x4"}, {AssetID: "asset://x5"}, {AssetID: "asset://x6"},
				}},
			}},
			// A 2x2 quadrant map (R14.9): axes + tints + plotted dots with long labels
			// at the corners (0,0)/(1,1) must stay on-canvas.
			scene.SceneSlide{ID: fmt.Sprintf("quadrant-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Quadrant{
					AxisX:     scene.QuadrantAxis{LowLabel: longText, HighLabel: "High effort"},
					AxisY:     scene.QuadrantAxis{LowLabel: "Low", HighLabel: longText},
					Quadrants: [4]scene.QuadrantCell{{Title: longText, Fill: &accent}, {Title: "TR"}, {Title: "BL"}, {Title: "BR"}},
					Items:     []scene.QuadrantItem{{X: 0, Y: 0, Label: longText}, {X: 1, Y: 1, Label: longText, AccentIndex: 2}, {X: 0.5, Y: 0.5, Label: "mid"}},
				},
			}},
			// An enriched testimonial (R14.5): oversized quote mark + a long multi-line
			// quote + structured attribution. The mark, text, and strip must stay on-canvas.
			scene.SceneSlide{ID: fmt.Sprintf("testimonial-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Quote{Text: rt(longText), Mark: true, AttributionName: "A very long attributed name here", AttributionRole: "Chief Revenue Officer", AttributionCompany: "Globex International Holdings"},
			}},
			// A fully-styled comparison matrix (R14.3): header band + grouped header
			// row + zebra + highlighted column + row labels, with long wrapping cells.
			// The table graphic frame must stay on-canvas; header text must contrast.
			scene.SceneSlide{ID: fmt.Sprintf("matrix-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Table{
					Headers: []scene.RichText{rt("Capability under evaluation"), rt("Free"), rt("Pro"), rt("Enterprise")},
					Rows: [][]scene.RichText{
						{rt(longText), rt("—"), rt("Included"), rt("Included")},
						{rt(longText), rt("—"), rt("99.9%"), rt("99.99%")},
					},
					Style: &scene.TableStyle{
						HeaderFill: true, Zebra: true, HighlightCol: 4, RowLabelCol: true,
						HeaderGroups: []scene.HeaderGroup{{Label: "Plan", Span: 1}, {Label: "Paid tiers", Span: 3}},
					},
				},
			}},
		)
	}
	return scene.Scene{Slides: slides}
}

var offRe = regexp.MustCompile(`<a:off x="(-?\d+)" y="(-?\d+)"/>`)
var extRe = regexp.MustCompile(`<a:ext cx="(\d+)" cy="(\d+)"/>`)

// TestAdversarial_AllBoxesOnCanvas is the R11.12 invariant: across every slide of
// the hostile fixture, every emitted shape/text box lies fully within the slide
// canvas (no off-slide / clipped content). A component that reintroduces a
// fixed-size assumption and draws off-slide fails here.
func TestAdversarial_AllBoxesOnCanvas(t *testing.T) {
	sc := adversarialScene()
	data, _ := render(t, sc)
	slideW, slideH := pptx.New().SlideSize()

	for i := range sc.Slides {
		part := fmt.Sprintf("ppt/slides/slide%d.xml", i+1)
		xml := zipPart(t, data, part)
		offs := offRe.FindAllStringSubmatch(xml, -1)
		exts := extRe.FindAllStringSubmatch(xml, -1)
		if len(offs) == 0 {
			t.Errorf("%s (slide %q): no shapes emitted", part, sc.Slides[i].ID)
			continue
		}
		// Pair each <a:off> with the <a:ext> that follows it (xfrm order).
		n := len(offs)
		if len(exts) < n {
			n = len(exts)
		}
		for k := 0; k < n; k++ {
			x, _ := strconv.Atoi(offs[k][1])
			y, _ := strconv.Atoi(offs[k][2])
			cx, _ := strconv.Atoi(exts[k][1])
			cy, _ := strconv.Atoi(exts[k][2])
			if x < 0 || y < 0 || x+cx > slideW || y+cy > slideH {
				t.Errorf("%s (slide %q): box [x=%d y=%d cx=%d cy=%d] exceeds the %dx%d canvas",
					part, sc.Slides[i].ID, x, y, cx, cy, slideW, slideH)
			}
		}
	}
}

// TestAdversarial_Renders is a smoke guard that the hostile fixture renders without
// error and produces every slide (so the on-canvas suite has parts to check).
func TestAdversarial_Renders(t *testing.T) {
	sc := adversarialScene()
	data, stats := render(t, sc)
	if stats.Slides != len(sc.Slides) {
		t.Errorf("rendered %d slides, want %d", stats.Slides, len(sc.Slides))
	}
	if len(data) == 0 {
		t.Error("empty render")
	}
}

// TestAdversarial_Deterministic: the hostile fixture (every component + the
// safe-area clamp on its overflowing content) renders byte-identically across worker
// counts — the clamp is pure integer math, so it never depends on scheduling.
func TestAdversarial_Deterministic(t *testing.T) {
	sc := adversarialScene()
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if string(seq) != string(par) {
		t.Fatalf("adversarial: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}
