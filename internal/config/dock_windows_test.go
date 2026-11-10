package config

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestDockWindowsConfiguration(t *testing.T) {
	var cfg UserConfig
	if err := toml.Unmarshal([]byte(`[dock.windows]
show = "all"
style = "pill"
grouped = true
format = "{index}: {title} ({cwd})"
separator = ""
minimized_fg = "magenta"
minimized_bg = "surface"
`), &cfg); err != nil {
		t.Fatal(err)
	}
	w := cfg.Dock.Windows
	if !w.ShowAll() || !w.Grouped || w.TabStyle().Flat() || w.Separator != "" {
		t.Fatalf("window options were not decoded: %+v", w)
	}
	if w.MinimizedFg != "magenta" || w.MinimizedBg != "surface" {
		t.Fatalf("minimized style was not decoded: %+v", w)
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
