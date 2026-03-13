package app

import "testing"

func TestThemeByName_Found(t *testing.T) {
	th := ThemeByName("tokyonight")
	if th.Name != "tokyonight" {
		t.Fatalf("expected tokyonight, got %q", th.Name)
	}
}

func TestThemeByName_NotFound(t *testing.T) {
	th := ThemeByName("nonexistent")
	if th.Name != TokyoNight.Name {
		t.Fatalf("expected default tokyonight, got %q", th.Name)
	}
}

func TestThemeByName_CatppuccinMacchiato(t *testing.T) {
	th := ThemeByName("catppuccin-macchiato")
	if th.Name != "catppuccin-macchiato" {
		t.Fatalf("expected catppuccin-macchiato, got %q", th.Name)
	}
}
