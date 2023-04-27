package afile

import (
	"os"
	"path"
	"testing"
)

// TestExpandTilde tests the ExpandTilde function.
func TestExpandTilde(t *testing.T) {
	// Get current HOME directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	// Test cases.
	testCases := []struct {
		in  string
		out string
	}{
		{"~/foo", path.Join(homeDir, "foo")},
		{"~/", homeDir},
		{"~", "~"},
		{"~foo", "~foo"},
		{"foo", "foo"},
		{"", ""},
	}

	// Run test cases.
	for i, testCase := range testCases {
		out, err := ExpandTilde(testCase.in)
		if err != nil {
			t.Fatal(err)
		}
		if out != testCase.out {
			t.Errorf("Test case %d: Expected %q, got %q", i, testCase.out, out)
		}
	}
}
