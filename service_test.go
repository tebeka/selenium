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
		if len(args) != 5 {
			t.Errorf("args length = %d, want = %d", len(args), 5)
		} else {
			if args[0] != "Xvfb" {
				t.Errorf("args[0] = %s, want = %s", args[0], "Xvfb")
			}
			if args[1] != "-displayfd" {
				t.Errorf("args[1] = %s, want = %s", args[1], "-displayfd")
			}
			if args[2] != "3" {
				t.Errorf("args[2] = %s, want = %s", args[2], "3")
			}
			if args[3] != "-nolisten" {
				t.Errorf("args[3] = %s, want = %s", args[3], "-nolisten")
			}
			if args[4] != "tcp" {
				t.Errorf("args[4] = %s, want = %s", args[4], "tcp")
			}
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
		if len(args) != 8 {
			t.Errorf("args length = %d, want = %d", len(args), 8)
		} else {
			if args[0] != "Xvfb" {
				t.Errorf("args[0] = %s, want = %s", args[0], "Xvfb")
			}
			if args[1] != "-displayfd" {
				t.Errorf("args[1] = %s, want = %s", args[1], "-displayfd")
			}
			if args[2] != "3" {
				t.Errorf("args[2] = %s, want = %s", args[2], "3")
			}
			if args[3] != "-nolisten" {
				t.Errorf("args[3] = %s, want = %s", args[3], "-nolisten")
			}
			if args[4] != "tcp" {
				t.Errorf("args[4] = %s, want = %s", args[4], "tcp")
			}
			if args[5] != "-screen" {
				t.Errorf("args[5] = %s, want = %s", args[5], "-screen")
			}
			if args[6] != "0" {
				t.Errorf("args[6] = %s, want = %s", args[6], "0")
			}
			if args[7] != options.ScreenSize {
				t.Errorf("args[7] = %s, want = %s", args[7], options.ScreenSize)
			}
		}
	})
}
