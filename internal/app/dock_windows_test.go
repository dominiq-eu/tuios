package app

import (
	"strings"
	"testing"

	"github.com/Gaurav-Gosain/tuios/internal/config"
)

// Configuring the window component must not make the workspace tabs inherit
// its appearance or change the default minimized-only listing.
func TestDockWindowEntriesDefaultToMinimizedOnly(t *testing.T) {
	m := dockMinimizedOS(t)
	m.Windows[0].Minimized = false
	items := m.getDockItems()
	if len(items) != 1 || items[0].WindowIndex != 1 {
		t.Fatalf("unconfigured dock lists %+v, want only minimized bravo", items)
	}
	if m.dockWindowStyled() {
		t.Fatal("unconfigured window entries must retain legacy styling")
	}
}

// A window list contains both visible and minimized panes in this workspace.
// Entries remain in their original order when focus moves.
func TestDockWindowsShowAllKeepsStableEntryOrder(t *testing.T) {
	m := dockMinimizedOS(t)
	m.UserConfig = &config.UserConfig{
		Dock: config.DockConfig{Windows: config.DockWindowsConfig{Show: config.DockWindowsShowAll}},
	}
	m.Windows[0].Minimized = false
	m.FocusedWindow = 0
	before := m.getDockItems()
	if len(before) != 2 || before[0].WindowIndex != 0 || before[1].WindowIndex != 1 {
		t.Fatalf("all mode lists %+v, want alpha and bravo", before)
	}
	m.FocusedWindow = 1
	after := m.getDockItems()
	for i := range before {
		if before[i].WindowIndex != after[i].WindowIndex {
			t.Fatalf("focus change reordered window entries: before=%+v after=%+v", before, after)
		}
	}
}

// Renderer geometry, layout measurements and the hit rectangles must use the
// same separator width, including multi-column configured separators.
func TestDockWindowsFlatSeparatorKeepsClickableGeometry(t *testing.T) {
	m := dockMinimizedOS(t)
	m.UserConfig = &config.UserConfig{
		Dock: config.DockConfig{Windows: config.DockWindowsConfig{
			Show: config.DockWindowsShowAll, Style: config.DockTabStyleFlat,
			Separator: " | ", ActiveBg: "accent",
		}},
	}
	m.Windows[0].Minimized = false
	m.FocusedWindow = 0
	row, _ := m.renderDockString()
	if !strings.Contains(row, " | ") {
		t.Fatalf("dock did not draw the configured separator: %q", row)
	}
	if len(m.dockItemHits) != 2 {
		t.Fatalf("dock recorded %d hits, want two", len(m.dockItemHits))
	}
	for i, hit := range m.dockItemHits {
		if hit.WindowIndex != i {
			t.Errorf("entry %d routes to window %d", i, hit.WindowIndex)
		}
		if hit.X1-hit.X0 != m.getDockItems()[i].Width {
			t.Errorf("entry %d width %d differs from measured %d", i, hit.X1-hit.X0, m.getDockItems()[i].Width)
		}
		for x := hit.X0; x < hit.X1; x++ {
			if got := m.DockItemAt(x, hit.Y); got != hit.WindowIndex {
				t.Errorf("window entry %d column %d routes to %d", i, x, got)
			}
		}
	}
	left, right := m.dockItemHits[0], m.dockItemHits[1]
	if right.X0-left.X1 != len(" | ") {
		t.Errorf("separator has %d cells, want 3", right.X0-left.X1)
	}
	for x := left.X1; x < right.X0; x++ {
		if got := m.DockItemAt(x, left.Y); got != -1 {
			t.Errorf("separator column %d incorrectly targets pane %d", x, got)
		}
	}
}
