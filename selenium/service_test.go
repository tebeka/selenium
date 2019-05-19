package selenium

import (
	"fmt"
	"testing"
	"time"

	"github.com/BurntSushi/xgbutil"
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
	// Note on FrameBuffer and xgb.Conn:
	// There appears to be a race condition when closing a Conn instance before
	// a FrameBuffer instance.  A short sleep solves the problem.
	t.Run("Default behavior", func(t *testing.T) {
		// The default Xvfb screen size is "1280x1024x8".
		frameBuffer, err := NewFrameBuffer()
		if err != nil {
			t.Fatalf("Could not create frame buffer: %s", err.Error())
		}
		defer frameBuffer.Stop()

		if frameBuffer.Display == "" {
			t.Fatalf("frameBuffer.Display is empty")
		}

		d, err := xgbutil.NewConnDisplay(":" + frameBuffer.Display)
		if err != nil {
			t.Fatalf("could not connect to display %q: %s", frameBuffer.Display, err.Error())
		}
		defer time.Sleep(time.Second * 2)
		defer d.Conn().Close()
		s := d.Screen()
		if diff := cmp.Diff(1280, int(s.WidthInPixels)); diff != "" {
			t.Fatalf("args returned diff (-want/+got):\n%s", diff)
		}
		if diff := cmp.Diff(1024, int(s.HeightInPixels)); diff != "" {
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
	t.Run("With screen size", func(t *testing.T) {
		desiredWidth := 1024
		desiredHeight := 768
		desiredDepth := 24
		options := FrameBufferOptions{
			ScreenSize: fmt.Sprintf("%dx%dx%d", desiredWidth, desiredHeight, desiredDepth),
		}
		frameBuffer, err := NewFrameBufferWithOptions(options)
		if err != nil {
			t.Fatalf("Could not create frame buffer: %s", err.Error())
		}
		defer frameBuffer.Stop()

		if frameBuffer.Display == "" {
			t.Fatalf("frameBuffer.Display is empty")
		}

		d, err := xgbutil.NewConnDisplay(":" + frameBuffer.Display)
		if err != nil {
			t.Fatalf("could not connect to display %q: %s", frameBuffer.Display, err.Error())
		}
		defer time.Sleep(time.Second * 2)
		defer d.Conn().Close()
		s := d.Screen()
		if diff := cmp.Diff(desiredWidth, int(s.WidthInPixels)); diff != "" {
			t.Fatalf("args returned diff (-want/+got):\n%s", diff)
		}
		if diff := cmp.Diff(desiredHeight, int(s.HeightInPixels)); diff != "" {
			t.Fatalf("args returned diff (-want/+got):\n%s", diff)
		}
	})
}
