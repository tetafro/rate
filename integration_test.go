// +build integration

package rate

import (
	"fmt"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	rateLimit := float64(100)
	runTime := 10 * time.Second

	lim, err := NewLimiter("localhost:6379", "zkey", rateLimit)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	lim.Log = &testLogger{}

	start := time.Now()
	var allowed, denied int
	for {
		if lim.Allow() {
			allowed++
		} else {
			denied++
		}
		if time.Since(start) > runTime {
			break
		}
	}
	want := int(runTime.Seconds() * rateLimit)
	if allowed > want {
		t.Fatalf("Too many allowed events:\n  want <=%d \n  got %d", want, allowed)
	}
	fmt.Println("Results:")
	fmt.Println("  allowed:", allowed)
	fmt.Println("  denied:", denied)
}
