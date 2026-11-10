package config

import "testing"

func TestDockWindowsDefaultsPreserveMinimizedBehavior(t *testing.T) {
	var w DockWindowsConfig
	if w.ShowAll() {
		t.Fatal("an omitted dock.windows table must not list visible panes")
	}
	style := w.TabStyle()
	if !style.Flat() {
		t.Fatalf("the configured window style defaults to flat, got %q", style.Style)
	}
	if style.Underlined() {
		t.Fatal("flat window entries must not underline by default")
	}
	if style.ActiveBg != "" || style.InactiveBg != "" {
		t.Fatal("window styling defaults must follow the theme, not hardcode colors")
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

func TestDockWindowsUnknownStyleFallsBackToFlat(t *testing.T) {
	style := (DockWindowsConfig{Style: "unknown"}).TabStyle()
	if !style.Flat() {
		t.Fatalf("unknown style should use flat rendering, got %q", style.Style)
	}
}
