package selenium

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

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

func TestFrameBuffer(t *testing.T) {
	// Make sure that we are using our unit-test version of `exec.Command`.
	newExecCommand = fakeExecCommand

	t.Run("Default behavior", func(t *testing.T) {
		frameBuffer, err := NewFrameBuffer()
		if err != nil {
			t.Fatalf("Could not create frame buffer: %s", err.Error())
		}
		if frameBuffer.Display != "1" {
			t.Errorf("frameBuffer.Display = %s, want %s", frameBuffer.Display, "1")
		}
		args := frameBuffer.cmd.Args[3:]
		expectedArgs := []string{"Xvfb", "-displayfd", "3", "-nolisten", "tcp"}
		if diff := cmp.Diff(expectedArgs, args); diff != "" {
			t.Fatalf("args returned diff (-want/+got):\n%s", diff)
		}
	})
	t.Run("With screen size", func(t *testing.T) {
		options := FrameBufferOptions{
			ScreenSize: "1024x768x24",
		}
		frameBuffer, err := NewFrameBufferWithOptions(options)
		if err != nil {
			t.Fatalf("Could not create frame buffer: %s", err.Error())
		}
		if frameBuffer.Display != "1" {
			t.Errorf("frameBuffer.Display = %s, want %s", frameBuffer.Display, "1")
		}
		args := frameBuffer.cmd.Args[3:]
		expectedArgs := []string{"Xvfb", "-displayfd", "3", "-nolisten", "tcp", "-screen", "0", options.ScreenSize}
		if diff := cmp.Diff(expectedArgs, args); diff != "" {
			t.Fatalf("args returned diff (-want/+got):\n%s", diff)
		}
	})
	t.Run("With bad screen size", func(t *testing.T) {
		options := FrameBufferOptions{
			ScreenSize: "not a screen size",
		}
		_, err := NewFrameBufferWithOptions(options)
		if err == nil {
			t.Fatalf("Expected an error about the screen size")
		}
	})
}
