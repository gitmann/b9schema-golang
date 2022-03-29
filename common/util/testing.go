package util

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func CompareStrings(t *testing.T, testName string, gotStrings, wantStrings []string) bool {
	// Split strings into lines.
	gotLines := []string{}
	for _, line := range gotStrings {
		lines := strings.Split(line, "\n")
		gotLines = append(gotLines, lines...)
	}
	wantLines := []string{}
	for _, line := range wantStrings {
		lines := strings.Split(line, "\n")
		wantLines = append(wantLines, lines...)
	}

	if !reflect.DeepEqual(gotLines, wantLines) {
		maxLen := len(gotLines)
		if len(wantLines) > maxLen {
			maxLen = len(wantLines)
		}

		type diffStruct struct {
			got, want string
		}
		diff := []*diffStruct{}

		for i := 0; i < maxLen; i++ {
			newDiff := &diffStruct{}

			if i < len(gotLines) {
				newDiff.got = gotLines[i]
			}

			if i < len(wantLines) {
				newDiff.want = wantLines[i]
			}

			diff = append(diff, newDiff)
		}

		// Dump got and want lines.
		outLines := []string{}

		outLines = append(outLines, "***** GOT:")
		for i, newDiff := range diff {
			flag := " "
			if newDiff.got != newDiff.want {
				flag = ">"
			}

			outLines = append(outLines, fmt.Sprintf("%05d%s|%s", i, flag, newDiff.got))
		}

		outLines = append(outLines, "***** WANT:")
		for i, newDiff := range diff {
			flag := " "
			if newDiff.got != newDiff.want {
				flag = ">"
			}

			outLines = append(outLines, fmt.Sprintf("%05d%s|%s", i, flag, newDiff.want))
		}

		t.Errorf("TEST_FAIL %s\n%s", testName, strings.Join(outLines, "\n"))
		return false
	} else {
		t.Logf("TEST_OK %s", testName)
		return true
	}
}

func OutputErrStrings(t *testing.T, testName string, gotStrings []string, err error) {
	gotLines := []string{}
	for _, s := range gotStrings {
		gotLines = append(gotLines, strings.Split(s, "\n")...)
	}

	outLines := []string{}

	outLines = append(outLines, "***** GOT:")
	for i, got := range gotLines {
		outLines = append(outLines, fmt.Sprintf("%05d |%s", i, got))
	}

	t.Errorf("TEST_FAIL %s: %s\n%s", testName, err, strings.Join(outLines, "\n"))
}
