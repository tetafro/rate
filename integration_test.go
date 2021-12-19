//go:build integration
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

const testRedisVersion = "6.2"

func TestLimiterIntegration(t *testing.T) {
	ctx := context.Background()

	redisAddr, teardown, err := newTestRedis(ctx)
	if err != nil {
		t.Fatalf("Failed to init redis: %v", err)
	}
	defer teardown(ctx)

	rateLimit := float64(100)
	runTime := 10 * time.Second

	lim, err := NewLimiter(redisAddr, "zkey", rateLimit)
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

func newTestRedis(ctx context.Context) (string, func(context.Context) error, error) {
	req := tc.ContainerRequest{
		Image:        "redis:" + testRedisVersion,
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redis, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return "", nil, fmt.Errorf("setup environment: %v", err)
	}

	host, err := redis.Host(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("get redis ip: %v", err)
	}
	mport, err := redis.MappedPort(ctx, "6379")
	if err != nil {
		return "", nil, fmt.Errorf("get redis port: %v", err)
	}
	addr := fmt.Sprintf("%s:%s", host, mport.Port())

	return addr, redis.Terminate, nil
}
