package tuie2e

import (
	"testing"

	"github.com/Gaurav-Gosain/tuitest"
)

// TestScrollingWideFirstNewPaneIsRevealed: with scroll_column_width = 90, the
// first pane opened by hand in the scrolling layout landed flush against the
// right edge of the screen with most of it hidden, and every pane opened after
// it was shown whole.
//
// The strip starts with one column, 90 of the screen's 100 cells wide, at the
// left edge. A new pane goes in the column after it, at x=90, so 10 cells of
// it are on screen. The focus change that followed used the least-scroll rule,
// which moves the strip only for a column with nothing on screen, and left it
// there. The second pane's column starts at x=180, wholly off screen, so that
// rule did move the strip for it, which is why only the first one was cut off.
// It is the same fault as issue 1507, reached without a keyboard step: there
// the peek was the one a step leaves, here it is the one a wide first column
// leaves. 90d814ed fixed both by showing the whole column of a pane the same
// sync created.
//
// Each door opens the first pane a different way: the report's own steps (the
// layout picked from the palette, the 'n' key), and a session started in the
// scrolling layout from the config with the pane opened from the command line,
// which is where the report's after-new-window hook runs.
//
// NEGATIVE CONTROL: see NEGATIVE_CONTROLS.md, the wide first column.
func TestScrollingWideFirstNewPaneIsRevealed(t *testing.T) {
	const cols = 100
	for _, door := range []string{"palette and key", "config and cli"} {
		t.Run(door, func(t *testing.T) {
			base := t.TempDir()
			config := "[startup]\nopen_default_window = true\n"
			if door == "config and cli" {
				config += "tiled = true\nlayout = \"scrolling\"\n"
			}
			writeConfig(t, base, config+"\n[appearance]\nscroll_column_width = 90\n")
			term := startIn(t, base, startOpts{cols: cols, rows: 30, args: []string{"new", "strip"}})
			waitWindowCount(t, term, 1, "the default window")
			if door == "palette and key" {
				enableTiling(t, term)
				runPaletteRow(t, term, "scrolling", "Layout: scrolling (niri-style)", "Layout: scrolling (niri)")
				waitPaletteClosed(t, term, "after picking the scrolling layout")
			}
			start := stripOrder(waitForSettledGeometry(t, base, 1))
			if err := term.WaitStable(uiTimeout); err != nil {
				t.Fatalf("the screen never settled: %v", err)
			}
			railShot(t, term, "before-first-new-pane")

			// wholeOnScreen fails unless pane id is focused and wholly on screen.
			wholeOnScreen := func(id string, n int, when string) {
				t.Helper()
				waitStripFocus(t, base, id, when)
				rects := waitForSettledGeometry(t, base, n)
				if err := term.WaitStable(uiTimeout); err != nil {
					t.Fatalf("the screen never settled: %v", err)
				}
				if r := rectByID(t, rects, id); r.X < 0 || r.X+r.Width > cols {
					t.Errorf("%s: the new pane is not wholly on screen: %+v\n%s", when, r, term.Snapshot())
				}
			}
			openPane := func(n int) string {
				t.Helper()
				before := stripOrder(waitForSettledGeometry(t, base, n-1))
				if door == "palette and key" {
					if err := term.SendKeys(tuitest.Ctrl('b'), "c"); err != nil {
						t.Fatalf("open a pane: %v", err)
					}
				} else if out, err := tuiosCLI(t, base, "new-window", "-s", "strip"); err != nil {
					t.Fatalf("new-window: %v\n%s", err, out)
				}
				waitWindowCount(t, term, n, "the new pane")
				return newPaneIn(t, before, waitForSettledGeometry(t, base, n))
			}

			var first string
			if door == "palette and key" {
				// The report's key: 'n' in window management mode, where picking
				// the layout left the client.
				newWindow(t, term)
				first = newPaneIn(t, start, waitForSettledGeometry(t, base, 2))
			} else {
				first = openPane(2)
			}
			wholeOnScreen(first, 2, "the first new pane")
			railShot(t, term, "after-first-new-pane")

			// The second one was always shown whole; it still is.
			wholeOnScreen(openPane(3), 3, "the second new pane")
			railShot(t, term, "after-second-new-pane")
		})
	}
}
