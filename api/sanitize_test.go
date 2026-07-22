package api

import "testing"

func TestSanitizePublicTextRemovesPrivateDetails(t *testing.T) {
	input := "load https://private.example/data?token=abc failed at /home/runner/build: authorization=secret"
	got := sanitizePublicText(input)
	for _, forbidden := range []string{"private.example", "token=abc", "/home/runner", "authorization=secret"} {
		if contains := len(got) > 0 && stringContains(got, forbidden); contains {
			t.Fatalf("sanitized text still contains %q: %q", forbidden, got)
		}
	}
}

func stringContains(value, part string) bool {
	for index := 0; index+len(part) <= len(value); index++ {
		if value[index:index+len(part)] == part {
			return true
		}
	}
	return false
}
