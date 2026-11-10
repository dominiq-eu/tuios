package app

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/Gaurav-Gosain/tuios/internal/config"
)

// Window tests cover the configured component's observable behavior. The
// legacy minimized dock, workspace tabs and sidebar title debounce are tested
// independently in their respective packages.
func dockWindowsOS(t *testing.T, windows config.DockWindowsConfig) *OS {
	t.Helper()
	m := dockMinimizedOS(t)
	m.UserConfig = &config.UserConfig{Dock: config.DockConfig{Windows: windows}}
	return m
}

func assertDockWindowHits(t *testing.T, m *OS, windows ...int) {
	t.Helper()
	if len(m.dockItemHits) != len(windows) {
		t.Fatalf("dock hit targets = %d, want %d", len(m.dockItemHits), len(windows))
	}
	for i, hit := range m.dockItemHits {
		if hit.WindowIndex != windows[i] {
			t.Errorf("hit %d selects window %d, want %d", i, hit.WindowIndex, windows[i])
		}
		if hit.X0 >= hit.X1 {
			t.Fatalf("hit %d is empty: %+v", i, hit)
		}
		for x := hit.X0; x < hit.X1; x++ {
			if got := m.DockItemAt(x, hit.Y); got != windows[i] {
				t.Errorf("column %d of entry %d selects window %d", x, i, got)
			}
		}
		if i > 0 {
			previous := m.dockItemHits[i-1]
			for x := previous.X1; x < hit.X0; x++ {
				if got := m.DockItemAt(x, hit.Y); got != -1 {
					t.Errorf("separator column %d selects window %d", x, got)
				}
			}
		}
	}
}

func TestDockWindowsSelection(t *testing.T) {
	m := dockMinimizedOS(t)
	m.Windows[0].Minimized = false
	if items := m.getDockItems(); len(items) != 1 || items[0].WindowIndex != 1 {
		t.Fatalf("without configuration, expected only minimized window 1, got %+v", items)
	}
	if m.dockWindowStyled() {
		t.Fatal("unconfigured windows must retain the legacy appearance")
	}

	m.UserConfig = &config.UserConfig{Dock: config.DockConfig{
		Windows: config.DockWindowsConfig{Show: config.DockWindowsShowAll},
	}}
	other := newTestWindow(t, "other-workspace", 60, 20)
	other.Workspace = 2
	scratch := newTestWindow(t, "scratch", 60, 20)
	scratch.Workspace, scratch.IsScratch = 1, true
	popup := newTestWindow(t, "popup", 60, 20)
	popup.Workspace, popup.IsPopup = 1, true
	m.Windows = append(m.Windows, other, scratch, popup)

	for _, focused := range []int{0, 1} {
		m.FocusedWindow = focused
		items := m.getDockItems()
		if len(items) != 2 || items[0].WindowIndex != 0 || items[1].WindowIndex != 1 {
			t.Fatalf("focus %d: show=all should list ordinary current-workspace windows in stable order, got %+v", focused, items)
		}
	}
}

func TestDockWindowsLabels(t *testing.T) {
	m := dockWindowsOS(t, config.DockWindowsConfig{
		Show: config.DockWindowsShowAll,
		Format: "({index}) {title} @ {cwd}",
	})
	m.Windows[0].Minimized = false
	m.Windows[0].CustomName = ""
	m.Windows[0].ForegroundCmd = "vim"
	m.Windows[0].ShellPgid = 0
	m.Windows[0].Cwd = "/home/user/project"

	m.Windows[1].CustomName = "notes"
	m.Windows[1].Host = "remote"
	m.Windows[1].ShellPgid = 0
	m.Windows[1].Cwd = "/srv/notes"

	items := m.getDockItems()
	for i, want := range []string{
		"(1) vim @ /home/user/project",
		"(2) notes @ /srv/notes",
	} {
		if got := strings.TrimSpace(items[i].Label); got != want {
			t.Errorf("window %d label = %q, want %q", i, got, want)
		}
	}
}

func TestDockWindowsRendering(t *testing.T) {
	for _, tc := range []struct {
		name, style, separator string
		grouped, caps          bool
		want                   func(left, right string) string
	}{
		{"pill", config.DockTabStylePill, "", false, false,
			func(l, r string) string { return l + " 1 " + r + l + " 2 " + r }},
		{"flat", config.DockTabStyleFlat, "", false, false,
			func(_, _ string) string { return " 1  2 " }},
		{"flat/caps", config.DockTabStyleFlat, "|", false, true,
			func(l, r string) string { return l + " 1 " + r + "|" + l + " 2 " + r }},
		{"grouped/pill", config.DockTabStylePill, " ", true, true,
			func(l, r string) string { return l + " 1 2 " + r }},
		{"grouped/flat", config.DockTabStyleFlat, "|", true, true,
			func(l, r string) string { return l + " 1|2 " + r }},
		{"grouped/no-caps", config.DockTabStyleFlat, "", true, false,
			func(_, _ string) string { return " 12 " }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := config.DockWindowsConfig{
				Show: config.DockWindowsShowAll, Format: "{index}",
				Style: tc.style, Grouped: tc.grouped, Caps: tc.caps,
				Separator: tc.separator,
			}
			if tc.name == "flat/caps" {
				// Inactive flat tabs without a background deliberately draw
				// spaces instead of coloured cap glyphs.
				options.InactiveBg = "panel"
			}
			m := dockWindowsOS(t, options)
			m.Settings.DockPillCaps = true
			m.Windows[0].Minimized = false
			m.FocusedWindow = 0

			row := dockRow(t, m)
			left, right := m.Settings.GetDockWorkspaceCapLeft(), m.Settings.GetDockWorkspaceCapRight()
			want := tc.want(left, right)
			if !strings.Contains(row, want) {
				t.Fatalf("expected window strip %q in rendered row %q", want, row)
			}
			assertDockWindowHits(t, m, 0, 1)
		})
	}
}

func TestDockWindowsInteraction(t *testing.T) {
	// Use a multicolumn separator and caps to check the boundary between
	// clickable entries against the actual rendered terminal columns.
	m := dockWindowsOS(t, config.DockWindowsConfig{
		Show: config.DockWindowsShowAll, Format: "{index}",
		Grouped: true, Style: config.DockTabStylePill,
		Caps: true, Separator: " • ",
	})
	m.Settings.DockPillCaps = true
	m.Windows[0].Minimized = false
	m.FocusedWindow = 0

	row := dockRow(t, m)
	left, right := m.Settings.GetDockWorkspaceCapLeft(), m.Settings.GetDockWorkspaceCapRight()
	strip := left + " 1 • 2 " + right
	start := strings.Index(row, strip)
	if start < 0 {
		t.Fatalf("expected window strip %q in rendered row %q", strip, row)
	}
	x0 := lipgloss.Width(row[:start])
	firstEnd := x0 + lipgloss.Width(left+" 1")
	secondStart := firstEnd + lipgloss.Width(" • ")
	secondEnd := x0 + lipgloss.Width(strip)
	assertDockWindowHits(t, m, 0, 1)
	if first, second := m.dockItemHits[0], m.dockItemHits[1]; first.X0 != x0 ||
		first.X1 != firstEnd || second.X0 != secondStart || second.X1 != secondEnd {
		t.Fatalf("rendered cells %d..%d and %d..%d do not match hit rectangles %+v",
			x0, firstEnd, secondStart, secondEnd, m.dockItemHits)
	}
	if got := lipgloss.Width(row); got > m.Width {
		t.Fatalf("rendered dock width %d exceeds %d", got, m.Width)
	}
}

func TestDockWindowsOverflow(t *testing.T) {
	m := dockCrowdedOS(t, 90, 5, 6)
	m.Settings.DockPillCaps = true
	m.UserConfig = &config.UserConfig{Dock: config.DockConfig{
		Windows: config.DockWindowsConfig{
			Grouped: true, Caps: true, Style: config.DockTabStylePill,
			Format: "{index}", Separator: " ",
		},
	}}

	// Exercise a real partial strip, without tying the test to an exact
	// surrounding workspace-strip or notification width.
	for _, width := range []int{55, 65, 75, 85, 95} {
		m.Width = width
		layout := m.CalculateDockLayout()
		if layout.TruncatedCount == 0 || len(layout.VisibleItems) == 0 {
			continue
		}
		row := dockRow(t, m)
		if !strings.Contains(row, m.Settings.GetDockWorkspaceCapRight()+" ...") {
			t.Errorf("truncated group did not close before overflow marker: %q", row)
		}
		if got := lipgloss.Width(row); got > width {
			t.Errorf("dock row has %d cells on %d-column terminal", got, width)
		}
		windows := make([]int, len(layout.VisibleItems))
		for i, item := range layout.VisibleItems {
			windows[i] = item.WindowIndex
		}
		assertDockWindowHits(t, m, windows...)
		return
	}
	t.Fatal("did not reach a partially visible grouped window strip")
}
