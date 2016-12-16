package selenium

// ChromeOptions defines the Chrome-specific desired capabilities.
// See https://sites.google.com/a/chromium.org/chromedriver/capabilities
type ChromeOptions struct {
	// Args are the command-line arguments to pass to the Chrome binary, in
	// addition to the ChromeDriver-supplied ones.
	Args []string `json:"args,omitempty"`

	// TODO(minusnine): finish the rest of the ChromeOptions struct.
}
