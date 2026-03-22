package gui

import (
	"testing"
)

func TestParseEnvLines(t *testing.T) {
	tests := []struct {
		name string
		text string
		want map[string]string
	}{
		{"simple", "KEY=val", map[string]string{"KEY": "val"}},
		{"multiple", "A=1\nB=2", map[string]string{"A": "1", "B": "2"}},
		{"quoted value", `KEY="hello world"`, map[string]string{"KEY": "hello world"}},
		{"single quoted", "KEY='hello'", map[string]string{"KEY": "hello"}},
		{"value with equals", "KEY=a=b=c", map[string]string{"KEY": "a=b=c"}},
		{"comment lines", "# comment\nKEY=val", map[string]string{"KEY": "val"}},
		{"blank lines", "\nKEY=val\n\n", map[string]string{"KEY": "val"}},
		{"spaces around key", " KEY = val ", map[string]string{"KEY": "val"}},
		{"empty input", "", map[string]string{}},
		{"no equals sign", "KEYONLY", map[string]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEnvLines(tt.text)
			if len(got) != len(tt.want) {
				t.Errorf("parseEnvLines(%q) returned %d entries, want %d", tt.text, len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("parseEnvLines(%q)[%q] = %q, want %q", tt.text, k, got[k], v)
				}
			}
		})
	}
}

func TestParseLines(t *testing.T) {
	tests := []struct {
		text string
		want []string
	}{
		{"a\nb\nc", []string{"a", "b", "c"}},
		{"  a  \n  b  ", []string{"a", "b"}},
		{"\n\n", nil},
		{"single", []string{"single"}},
		{"", nil},
	}
	for _, tt := range tests {
		got := parseLines(tt.text)
		if len(got) != len(tt.want) {
			t.Errorf("parseLines(%q) = %v, want %v", tt.text, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parseLines(%q)[%d] = %q, want %q", tt.text, i, got[i], tt.want[i])
			}
		}
	}
}

func TestStripQuotes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{`"mixed'`, `"mixed'`},
		{"noquotes", "noquotes"},
		{`""`, ""},
		{`"a"`, "a"},
		{`x`, "x"},
	}
	for _, tt := range tests {
		got := stripQuotes(tt.input)
		if got != tt.want {
			t.Errorf("stripQuotes(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLocaleLabelFromCode(t *testing.T) {
	got := localeLabelFromCode("ja_JP.UTF-8")
	want := "Japanese (ja_JP.UTF-8)"
	if got != want {
		t.Errorf("localeLabelFromCode(ja_JP.UTF-8) = %q, want %q", got, want)
	}
}

func TestLocaleLabelFromCodeUnknown(t *testing.T) {
	got := localeLabelFromCode("xx_XX.UTF-8")
	if got != "xx_XX.UTF-8" {
		t.Errorf("unknown code should return itself, got %q", got)
	}
}

func TestLocaleCodeFromLabel(t *testing.T) {
	got := localeCodeFromLabel("Japanese (ja_JP.UTF-8)")
	if got != "ja_JP.UTF-8" {
		t.Errorf("localeCodeFromLabel = %q, want ja_JP.UTF-8", got)
	}
}

func TestLocaleCodeFromLabelUnknown(t *testing.T) {
	got := localeCodeFromLabel("Unknown Label")
	if got != "Unknown Label" {
		t.Errorf("unknown label should return itself, got %q", got)
	}
}

func TestLocaleOptionsList(t *testing.T) {
	opts := localeOptionsList()
	if len(opts) == 0 {
		t.Fatal("localeOptionsList returned empty")
	}
	if opts[0] != "System Default" {
		t.Errorf("first option should be System Default, got %q", opts[0])
	}
	if len(opts) != len(localeLabels)+1 {
		t.Errorf("localeOptionsList has %d items, want %d", len(opts), len(localeLabels)+1)
	}
}
