package rate

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

func TestNewLimiter(t *testing.T) {
	_, err := NewLimiter("localhost:9999", "zkey", 10)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestLimiter_Init(t *testing.T) {
	conn := redigomock.NewConn()
	p := &redis.Pool{
		Dial: func() (redis.Conn, error) { return conn, nil },
	}

	lim := Limiter{
		Window: time.Second,
		Limit:  10,
		Pool:   p,
		Key:    "zkey",
		Log:    defaultLogger(),
	}

	t.Run("success", func(t *testing.T) {
		dig := "abc"
		conn.Command("SCRIPT", "LOAD", script).Expect(dig)

		err := lim.Init()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if lim.Digest != dig {
			t.Fatalf("Invalid digest\n  want: %s\n  got: %s", dig, lim.Digest)
		}
	})

	t.Run("redis error", func(t *testing.T) {
		conn.Command("SCRIPT", "LOAD", script).ExpectError(errors.New("fail"))

		err := lim.Init()
		if err == nil {
			t.Fatal("Expected error, go nil")
		}
	})
}

func TestLimiter_Allow(t *testing.T) {
	conn := redigomock.NewConn()
	p := &redis.Pool{
		Dial: func() (redis.Conn, error) { return conn, nil },
	}

	lim := Limiter{
		Window: time.Second,
		Limit:  10,
		Pool:   p,
		Key:    "zkey",
		Digest: "abc",
		Log:    &testLogger{},
	}

	t.Run("success (allow)", func(t *testing.T) {
		conn.Command(
			"EVALSHA", lim.Digest, 1,
			lim.Key, lim.Window.Nanoseconds(), lim.Limit,
			redigomock.NewAnyData(),
		).Expect(int64(1))

		if !lim.Allow() {
			t.Fatal("Invalid result:\n  want true\n  got false")
		}
		if buf := lim.Log.(*testLogger).buffer; buf != "" {
			t.Fatalf("Unexpected log entry: %s", buf)
		}
	})

	t.Run("success (deny)", func(t *testing.T) {
		conn.Command(
			"EVALSHA", lim.Digest, 1,
			lim.Key, lim.Window.Nanoseconds(), lim.Limit,
			redigomock.NewAnyData(),
		).Expect(int64(0))

		if lim.Allow() {
			t.Fatal("Invalid result:\n  want false\n  got true")
		}
		if buf := lim.Log.(*testLogger).buffer; buf != "" {
			t.Fatalf("Unexpected log entry: %s", buf)
		}
	})

	t.Run("redis error", func(t *testing.T) {
		conn.Command(
			"EVALSHA", lim.Digest, 1,
			lim.Key, lim.Window.Nanoseconds(), lim.Limit,
			redigomock.NewAnyData(),
		).ExpectError(errors.New("fail"))

		if !lim.Allow() {
			t.Fatal("Invalid result:\n  want true\n  got false")
		}
		want := "Redis request failed: fail"
		if buf := lim.Log.(*testLogger).buffer; buf != want {
			t.Fatalf("Invalid log entry\n  want: %s\n  got: %s", want, buf)
		}
	})
}

type testLogger struct {
	buffer string
}

func (l *testLogger) Printf(msg string, args ...interface{}) {
	l.buffer = fmt.Sprintf(msg, args...)
}
