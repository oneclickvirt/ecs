//go:build !windows

package runner

import (
	"reflect"
	"testing"
)

func TestParseProcStatParentHandlesSpacesAndParentheses(t *testing.T) {
	parent, ok := parseProcStatParent([]byte("123 (benchmark worker) name) S 42 123 123 0"))
	if !ok || parent != 42 {
		t.Fatalf("parseProcStatParent() = %d, %t", parent, ok)
	}
}

func TestDescendantsFromParentsReturnsDeepestFirst(t *testing.T) {
	parents := map[int]int{11: 10, 12: 10, 21: 11, 22: 11, 31: 21, 99: 1}
	got := descendantsFromParents(10, parents)
	want := []int{31, 21, 22, 11, 12}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("descendantsFromParents() = %v, want %v", got, want)
	}
}
