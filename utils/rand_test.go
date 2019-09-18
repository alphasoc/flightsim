package utils

import (
	"strings"
	"testing"
)

func TestRandString(t *testing.T) {
	s := RandString(100)
	if len(s) != 100 {
		t.Fatal(len(s))
	}
	for _, c := range []byte(s) {
		if strings.IndexByte(letterBytes, c) < 0 {
			t.Errorf("invalid byte: 0x%x", c)
		}
	}
}
