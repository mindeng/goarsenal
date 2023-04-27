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

// TestReadLines tests the ReadLines function.
func TestReadLines(t *testing.T) {
	// Test cases.
	testCases := []struct {
		in  string
		out []string
	}{
		{"testdata/empty.txt", []string{}},
		{"testdata/one.txt", []string{"one"}},
		{"testdata/two.txt", []string{"one", "", "two"}},
		{"testdata/three.txt", []string{"one", "two", "three"}},
	}

	// Run test cases.
	for i, testCase := range testCases {
		out, err := ReadLines(testCase.in)
		if err != nil {
			t.Fatal(err)
		}
		if len(out) != len(testCase.out) {
			t.Errorf("Test case %d: Expected %d lines, got %d", i, len(testCase.out), len(out))
		}
		for j, line := range out {
			if line != testCase.out[j] {
				t.Errorf("Test case %d: Expected line %d to be %q, got %q", i, j, testCase.out[j], line)
			}
		}
	}
}
