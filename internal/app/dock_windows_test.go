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

func TestDockWindowTitlesMatchSidebarTerminals(t *testing.T) {
	m := dockMinimizedOS(t)
	m.UserConfig = &config.UserConfig{
		Dock: config.DockConfig{Windows: config.DockWindowsConfig{
			Show: config.DockWindowsShowAll,
			Format: "{index}: {title}",
		}},
	}
	m.Windows[0].Minimized = false
	m.Windows[0].CustomName = ""
	m.Windows[0].ForegroundCmd = "vim"
	m.Windows[1].CustomName = ""
	m.Windows[1].ForegroundCmd = ""
	m.Windows[1].SetTitle("")
	for i, w := range m.Windows {
		want := printableTitle(m.railTitleShown(w))
		if want == "" {
			want = "shell"
		}
		item := m.getDockItems()[i]
		if item.Name != want {
			t.Errorf("window %d dock name = %q, sidebar terminal name = %q", i, item.Name, want)
		}
		if !strings.Contains(item.Label, want) {
			t.Errorf("window %d label %q does not include sidebar title %q", i, item.Label, want)
		}
	}
	if got := m.getDockItems()[0].Name; got != "vim" {
		t.Fatalf("foreground command is not the dock title: %q", got)
	}
	if got := m.getDockItems()[1].Name; got != "shell" {
		t.Fatalf("unnamed terminal lacks sidebar fallback: %q", got)
	}
}

func TestDockWindowTitleFollowsDebouncedSidebarTitle(t *testing.T) {
	m := dockMinimizedOS(t)
	m.UserConfig = &config.UserConfig{
		Dock: config.DockConfig{Windows: config.DockWindowsConfig{
			Show: config.DockWindowsShowAll, Format: "{title}",
		}},
	}
	m.Windows[0].CustomName = ""
	m.Windows[0].ForegroundCmd = ""
	m.Windows[0].SetTitle("old")
	m.updateRailTitles()
	m.Windows[0].SetTitle("new")
	if got := m.getDockItems()[0].Name; got != "old" {
		t.Fatalf("dock jumped ahead of debounced sidebar title: %q", got)
	}
	backdate(m, m.Windows[0].ID, railTitleDebounce)
	m.updateRailTitles()
	if got := m.getDockItems()[0].Name; got != "new" {
		t.Fatalf("dock did not follow updated sidebar title: %q", got)
	}
}

func TestDockWindowCWDUsesPaneDirectory(t *testing.T) {
	m := dockMinimizedOS(t)
	m.UserConfig = &config.UserConfig{
		Dock: config.DockConfig{Windows: config.DockWindowsConfig{
			Show: config.DockWindowsShowAll, Format: "{cwd}",
		}},
	}
	w := m.Windows[0]
	w.CustomName = ""
	w.ShellPgid = 0 // daemon panes may not have a process on the client
	for _, tc := range []struct {
		name, cwd, host, want string
	}{
		{"shell report", "/home/user/project", "", "/home/user/project"},
		{"daemon-managed pane", "/srv/app", "", "/srv/app"},
		{"remote pane", "/remote/repo", "build-host", "/remote/repo"},
		{"unknown directory", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w.Cwd, w.Host = tc.cwd, tc.host
			items := m.getDockItems()
			if len(items) < 1 {
				t.Fatal("missing window entry")
			}
			if got := items[0].CWD; got != tc.want {
				t.Fatalf("CWD = %q, want %q", got, tc.want)
			}
			if got := strings.TrimSpace(items[0].Label); got != tc.want {
				t.Fatalf("formatted label = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDockGroupedWindowsLayoutAndHitTargets(t *testing.T) {
	for _, tc := range []struct {
		name, style, separator string
		caps bool
	}{
		{"pill/no-caps", config.DockTabStylePill, "", false},
		{"pill/caps", config.DockTabStylePill, "", true},
		{"flat/no-caps", config.DockTabStyleFlat, "", false},
		{"flat/caps", config.DockTabStyleFlat, "", true},
		{"flat/space-separator", config.DockTabStyleFlat, " ", true},
		{"pill/space-separator", config.DockTabStylePill, " ", true},
		{"flat/separator", config.DockTabStyleFlat, "|", true},
		{"pill/separator", config.DockTabStylePill, "|", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := dockMinimizedOS(t)
			m.Settings.DockPillCaps = true
			m.UserConfig = &config.UserConfig{Dock: config.DockConfig{
				Windows: config.DockWindowsConfig{
					Show: config.DockWindowsShowAll, Grouped: true,
					Style: tc.style, Caps: tc.caps, Format: "{index}",
					Separator: tc.separator,
				},
			}}
			m.Windows[0].Minimized = false
			m.FocusedWindow = 0
			layout := m.CalculateDockLayout()
			if len(layout.VisibleItems) != 2 || layout.TruncatedCount != 0 {
				t.Fatalf("expected two complete window entries, got %+v", layout)
			}
			left, right := "", ""
			if tc.caps {
				left = m.Settings.GetDockWorkspaceCapLeft()
				right = m.Settings.GetDockWorkspaceCapRight()
			}
			// No implicit space: only the first label has the group's
			// opening padding. In particular, separator="" renders "12".
			want := left + " 1" + tc.separator + "2 " + right
			row := dockRow(t, m)
			if !strings.Contains(row, want) {
				t.Fatalf("grouped strip should draw %q, got:\n%s", want, row)
			}
			if len(m.dockItemHits) != 2 {
				t.Fatalf("expected 2 click targets, got %d", len(m.dockItemHits))
			}
			h0, h1 := m.dockItemHits[0], m.dockItemHits[1]
			if got := h0.X1 - h0.X0; got != layout.VisibleItems[0].Width+len([]rune(left)) {
				t.Errorf("first window spans %d cells, want %d", got, layout.VisibleItems[0].Width+len([]rune(left)))
			}
			if got := h1.X1 - h1.X0; got != layout.VisibleItems[1].Width+1+len([]rune(right)) {
				t.Errorf("last window spans %d cells, want %d", got, layout.VisibleItems[1].Width+1+len([]rune(right)))
			}
			if h1.X0-h0.X1 != len(tc.separator) {
				t.Errorf("separator width = %d, want %d", h1.X0-h0.X1, len(tc.separator))
			}
			if tc.separator == "" && h0.X1 != h1.X0 {
				t.Fatal("empty separator must make adjacent window hitboxes contiguous")
			}
			for i, hit := range m.dockItemHits {
				if hit.WindowIndex != i {
					t.Errorf("window %d hit targets window %d", i, hit.WindowIndex)
				}
				for x := hit.X0; x < hit.X1; x++ {
					if got := m.DockItemAt(x, hit.Y); got != i {
						t.Errorf("window %d column %d targets %d", i, x, got)
					}
				}
			}
			// Focus affects colors only; it must not reorder or resize entries.
			m.FocusedWindow = 1
			after := dockRow(t, m)
			if !strings.Contains(after, want) {
				t.Errorf("focus moved the group: %s", after)
			}
			for i, hit := range m.dockItemHits {
				if hit.WindowIndex != i {
					t.Errorf("focus change moved click target %d to %d", i, hit.WindowIndex)
				}
			}
		})
	}
}

func TestDockGroupedWindowOverflowStillClosesTheVisibleStrip(t *testing.T) {
	m := dockCrowdedOS(t, 90, 5, 6)
	m.Settings.DockPillCaps = true
	m.UserConfig = &config.UserConfig{Dock: config.DockConfig{
		Windows: config.DockWindowsConfig{
			Grouped: true, Caps: true, Style: config.DockTabStylePill,
			Format: "{index}",
		},
	}}
	for _, width := range []int{55, 65, 75, 85, 95} {
		m.Width = width
		layout := m.CalculateDockLayout()
		if layout.TruncatedCount == 0 || len(layout.VisibleItems) == 0 {
			continue
		}
		row := dockRow(t, m)
		if !strings.Contains(row, m.Settings.GetDockWorkspaceCapRight()+" ...") {
			t.Errorf("a truncated window strip did not close its visible group: %s", row)
		}
		if len(m.dockItemHits) != len(layout.VisibleItems) {
			t.Fatalf("%d visible entries, %d click targets", len(layout.VisibleItems), len(m.dockItemHits))
		}
		for _, h := range m.dockItemHits {
			for x := h.X0; x < h.X1; x++ {
				if m.DockItemAt(x, h.Y) != h.WindowIndex {
					t.Errorf("truncated entry at column %d did not target pane %d", x, h.WindowIndex)
				}
			}
		}
		return
	}
	t.Fatal("test did not reach a partially truncated group")
}
