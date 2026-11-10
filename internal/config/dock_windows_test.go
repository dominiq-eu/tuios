package config

import "testing"

func TestDockWindowsDefaultsPreserveMinimizedBehavior(t *testing.T) {
	var w DockWindowsConfig
	if w.ShowAll() {
		t.Fatal("an omitted dock.windows table must not list visible panes")
	}
	style := w.TabStyle()
	if style.Flat() {
		t.Fatalf("the configured window style defaults to pill, got %q", style.Style)
	}
	if !style.Underlined() {
		t.Fatal("pill window entries should follow the tab underline default")
	}
	if style.ActiveBg != "accent" || style.InactiveBg != "" {
		t.Fatalf("default pill should highlight the active pane on accent, not inactive ones: %+v", style)
	}
}

func TestDockWindowsStyleUsesTabSemantics(t *testing.T) {
	underline := false
	w := DockWindowsConfig{
		Show: DockWindowsShowAll, Style: DockTabStylePill, Separator: " | ",
		ActiveFg: "text", ActiveBg: "accent", InactiveFg: "dim",
		InactiveBg: "surface", ActiveBold: true, InactiveBold: true,
		ActiveUnderline: &underline, Caps: true,
	}
	if !w.ShowAll() {
		t.Fatal("show=all must list workspace panes")
	}
	style := w.TabStyle()
	if style.Flat() || style.Underlined() {
		t.Fatal("pill rendering and explicit underline override must be preserved")
	}
	if style.Separator != w.Separator || style.ActiveBg != w.ActiveBg ||
		style.InactiveBg != w.InactiveBg || style.ActiveBold != w.ActiveBold ||
		style.InactiveBold != w.InactiveBold || style.Caps != w.Caps {
		t.Fatalf("window tab-style adapter lost settings: %+v", style)
	}
}

func TestDockWindowsUnknownStyleFallsBackToPill(t *testing.T) {
	style := (DockWindowsConfig{Style: "unknown"}).TabStyle()
	if style.Flat() {
		t.Fatalf("unknown style should use pill rendering, got %q", style.Style)
	}
}

func TestDockWindowsFlatStyleAndFormat(t *testing.T) {
	w := DockWindowsConfig{Style: DockTabStyleFlat, Format: "{index} | {title} @ {cwd}"}
	if !w.TabStyle().Flat() || w.TabStyle().Underlined() {
		t.Fatal("flat window entries should not be underlined by default")
	}
	if got := FormatDockWindowLabel(w.Format, "editor", 3, "/work"); got != "3 | editor @ /work" {
		t.Fatalf("format expanded to %q", got)
	}
	if got := FormatDockWindowLabel("", "editor", 3, "/work"); got != "3:editor" {
		t.Fatalf("default label expanded to %q", got)
	}
	if got := FormatDockWindowLabel("", "", 3, "/work"); got != "3" {
		t.Fatalf("nameless default label expanded to %q", got)
	}
}
