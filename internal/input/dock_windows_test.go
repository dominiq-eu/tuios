package input

import (
	"testing"

	"github.com/Gaurav-Gosain/tuios/internal/app"
	"github.com/Gaurav-Gosain/tuios/internal/config"
	"github.com/Gaurav-Gosain/tuios/internal/terminal"
)

// The app package tests drawn hit rectangles. This one crosses the input
// boundary: clicking a visible window focuses it; clicking a minimized one
// restores it, without requiring knowledge of either window's screen box.
func TestDockWindowsClickFocusesOrRestores(t *testing.T) {
	separator := " "
	cfg := config.DefaultConfig()
	cfg.Dock.Windows = config.DockWindowsConfig{
		Show: config.DockWindowsShowAll,
		Grouped: true,
		Style: config.DockTabStylePill,
		Format: "{index}",
		Separator: &separator,
	}
	o := app.NewOS(app.OSOptions{
		UserConfig:      cfg,
		KeybindRegistry: config.NewKeybindRegistry(cfg),
	})
	o.Width, o.Height = 160, 40
	o.EffectiveWidth, o.EffectiveHeight = 160, 40
	o.Windows = []*terminal.Window{
		{ID: "visible", CustomName: "editor", Workspace: 1, Width: 50, Height: 20},
		{ID: "minimized", CustomName: "logs", Workspace: 1, Minimized: true,
			PreMinimizeWidth: 50, PreMinimizeHeight: 20},
	}
	o.CurrentWorkspace, o.FocusedWindow = 1, 0

	clickWindow := func(index int) {
		t.Helper()
		_ = o.View() // records dock hit targets for this frame
		y := o.GetDockbarContentYPosition()
		for x := 0; x < o.Width; x++ {
			if o.DockItemAt(x, y) == index {
				o, _ = handleMouseClick(clickAt(x, y), o)
				return
			}
		}
		t.Fatalf("dock did not render a click target for window %d", index)
	}

	clickWindow(1)
	if o.Windows[1].Minimized || o.FocusedWindow != 1 {
		t.Fatalf("minimized window was not restored and focused: minimized=%v focus=%d",
			o.Windows[1].Minimized, o.FocusedWindow)
	}

	clickWindow(0)
	if o.FocusedWindow != 0 {
		t.Fatalf("clicking a visible dock window focused %d, want 0", o.FocusedWindow)
	}
}
