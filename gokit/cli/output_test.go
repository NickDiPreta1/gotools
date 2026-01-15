package cli

import "testing"

func TestColorize(t *testing.T) {
	tests := []struct {
		name          string
		colorsEnabled bool
		color         string
		text          string
		want          string
	}{
		{
			name:          "colors enabled",
			colorsEnabled: true,
			color:         Red,
			text:          "error",
			want:          Red + "error" + Reset,
		},
		{
			name:          "colors disabled",
			colorsEnabled: false,
			color:         Red,
			text:          "error",
			want:          "error",
		},
		{
			name:          "empty text with colors",
			colorsEnabled: true,
			color:         Green,
			text:          "",
			want:          Green + "" + Reset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetColorsEnabled(tt.colorsEnabled)
			got := Colorize(tt.color, tt.text)
			if got != tt.want {
				t.Errorf("Colorize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSuccess(t *testing.T) {
	SetColorsEnabled(true)
	got := Success("ok")
	want := Green + "ok" + Reset
	if got != want {
		t.Errorf("Success() = %q, want %q", got, want)
	}
}

func TestError(t *testing.T) {
	SetColorsEnabled(true)
	got := Error("fail")
	want := Red + "fail" + Reset
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestWarning(t *testing.T) {
	SetColorsEnabled(true)
	got := Warning("warn")
	want := Yellow + "warn" + Reset
	if got != want {
		t.Errorf("Warning() = %q, want %q", got, want)
	}
}

func TestInfo(t *testing.T) {
	SetColorsEnabled(true)
	got := Info("info")
	want := Cyan + "info" + Reset
	if got != want {
		t.Errorf("Info() = %q, want %q", got, want)
	}
}

func TestSetColorsEnabled(t *testing.T) {
	SetColorsEnabled(false)
	if got := Success("test"); got != "test" {
		t.Errorf("Expected plain text when colors disabled, got %q", got)
	}

	SetColorsEnabled(true)
	if got := Success("test"); got == "test" {
		t.Errorf("Expected colored text when colors enabled, got plain text")
	}
}
