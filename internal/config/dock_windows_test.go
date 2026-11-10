package config

import "testing"

func TestDockWindowsConfiguration(t *testing.T) {
	cfg, err := ParseUserConfig([]byte(`[dock.windows]
show = "all"
style = "pill"
grouped = true
format = "{index}: {title} ({cwd})"
separator = ""
minimized_fg = "magenta"
minimized_bg = "surface"
`))
	if err != nil {
		t.Fatal(err)
	}
	w := cfg.Dock.Windows
	if !w.ShowAll() || !w.Grouped || w.TabStyle().Flat() || w.SeparatorText() != "" {
		t.Fatalf("window options were not decoded: %+v", w)
	}
	if w.MinimizedFg != "magenta" || w.MinimizedBg != "surface" {
		t.Fatalf("minimized style was not decoded: %+v", w)
	}
	empty, err := ParseUserConfig([]byte("[dock.windows]\nseparator = \"\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !empty.Dock.Windows.Defined || empty.Dock.Windows.Separator == nil || empty.Dock.Windows.SeparatorText() != "" {
		t.Fatalf("an explicit empty separator must activate the windows table: %+v", empty.Dock.Windows)
	}
	serialized, err := MarshalUserConfig(empty)
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := ParseUserConfig(serialized)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Dock.Windows.Separator == nil || reloaded.Dock.Windows.SeparatorText() != "" {
		t.Fatalf("config round trip dropped the explicitly empty separator: %+v", reloaded.Dock.Windows)
	}
	omitted, err := ParseUserConfig([]byte("[dock.windows]\nstyle = \"flat\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !omitted.Dock.Windows.Defined || omitted.Dock.Windows.Separator != nil || omitted.Dock.Windows.SeparatorText() != " " {
		t.Fatalf("omitting separator must use one space: %+v", omitted.Dock.Windows)
	}
	custom, err := ParseUserConfig([]byte("[dock.windows]\nseparator = \" | \"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if custom.Dock.Windows.SeparatorText() != " | " {
		t.Fatalf("custom separator should survive parsing: %+v", custom.Dock.Windows)
	}
	absent, err := ParseUserConfig([]byte("[dock.tabs]\nstyle = \"flat\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if absent.Dock.Windows.Defined {
		t.Fatal("settings for workspace tabs must not activate dock window styling")
	}
}

func TestDockWindowsDefaultsAndStyle(t *testing.T) {
	var w DockWindowsConfig
	if w.ShowAll() || w.Grouped || w.TabStyle().Flat() {
		t.Fatalf("default dock style must be ungrouped pill with minimized-only entries: %+v", w)
	}
	if !w.TabStyle().Underlined() || w.TabStyle().ActiveBg != "accent" {
		t.Fatalf("pill did not retain active styling defaults: %+v", w.TabStyle())
	}
	if w.TabStyle().InactiveFg != "text" {
		t.Fatalf("visible unfocused windows should use normal text: %+v", w.TabStyle())
	}
	if !(DockWindowsConfig{Style: DockTabStyleFlat}).TabStyle().Flat() {
		t.Fatal("flat window style was ignored")
	}
	if (DockWindowsConfig{Style: DockTabStyleFlat}).TabStyle().Underlined() {
		t.Fatal("flat windows should not be underlined by default")
	}
	if (DockWindowsConfig{Style: "unknown"}).TabStyle().Flat() {
		t.Fatal("invalid styles must fall back to pill")
	}
}

func TestDockWindowLabelFormatting(t *testing.T) {
	for _, tc := range []struct {
		name, format, title, cwd, want string
		index int
	}{
		{"current directory", "{index}: {title} ({cwd})", "editor", "/home/user/project", "3: editor (project)", 3},
		{"trailing slash", "{cwd}", "", "/work/repo/", "repo", 1},
		{"remote Windows path", "{cwd}", "", `C:\work\tools`, "tools", 1},
		{"root", "{cwd}", "", "/", "/", 1},
		{"unknown directory", "{cwd}", "", "", "", 1},
		{"default title", "", "editor", "/work", "3:editor", 3},
		{"default number", "", "", "/work", "3", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatDockWindowLabel(tc.format, tc.title, tc.index, tc.cwd)
			if got != tc.want {
				t.Fatalf("label = %q, want %q", got, tc.want)
			}
		})
	}
}
