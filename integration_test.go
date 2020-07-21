// +build integration

package rate

import (
	"context"
	"fmt"
	"testing"
	"time"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestLimiterIntegration(t *testing.T) {
	ctx := context.Background()
	redis, err := setup(ctx)
	if err != nil {
		t.Fatalf("Failed to setup environment: %v", err)
	}
	defer teardown(ctx, redis)

	rateLimit := float64(100)
	runTime := 10 * time.Second

	lim, err := NewLimiter("localhost:6379", "zkey", rateLimit)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

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

func setup(ctx context.Context) (tc.Container, error) {
	req := tc.ContainerRequest{
		Image:        "redis:3.2",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	return tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func teardown(ctx context.Context, c tc.Container) {
	c.Terminate(ctx)
}
