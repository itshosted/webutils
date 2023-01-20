package str

import (
	"math/rand"
	"testing"
	"time"
)

func TestRandText(t *testing.T) {
	uniq := make(map[string]bool)
	for i := 0; i < 5000; i++ {
		key := RandText(32)
		if _, ok := uniq[key]; !ok {
			uniq[key] = true
		} else {
			t.Fatalf("RandText uniq fail in round %d", i)
		}
	}
}

func TestRandText2(t *testing.T) {
	src := rand.NewSource(time.Now().UnixNano())
	uniq := make(map[string]bool)
	for i := 0; i < 5000; i++ {
		key := RandText2(src, 32)
		if _, ok := uniq[key]; !ok {
			uniq[key] = true
		} else {
			t.Fatalf("RandText2 uniq fail in round %d", i)
		}
	}
}
