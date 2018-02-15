package selenium

import "testing"

func TestIsDisplay(t *testing.T) {
	tests := []struct {
		desc  string
		in    string
		valid bool
	}{
		{
			desc:  "valid with just display",
			in:    "2",
			valid: true,
		},
		{
			desc:  "valid with display and screen",
			in:    "2.5",
			valid: true,
		},
		{
			desc:  "invalid with non-numeric display",
			in:    "a",
			valid: false,
		},
		{
			desc:  "invalid with non-numeric display and screen",
			in:    "a.5",
			valid: false,
		},
		{
			desc:  "invalid with display and non-numeric screen",
			in:    "2.b",
			valid: false,
		},
		{
			desc:  "invalid with display and blank screen",
			in:    "2.",
			valid: false,
		},
		{
			desc:  "invalid with blank display and screen",
			in:    ".3",
			valid: false,
		},
		{
			desc:  "invalid with blank display and blank screen",
			in:    ".",
			valid: false,
		},
		{
			desc:  "blank string is invalid",
			in:    "",
			valid: false,
		},
		{
			desc:  "malformed input",
			in:    "2.5.7",
			valid: false,
		},
	}

	for _, test := range tests {
		if got, want := isDisplay(test.in), test.valid; got != want {
			t.Errorf("%s: isDisplay = %t, want %t", test.desc, got, want)
		}
	}
}
