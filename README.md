# Rate limiter

Dead simple distributed redis-based rate limiter.

Uses redis [ZSET](https://redis.io/topics/data-types#sorted-sets) structure for
storing events.

Basically the algorithm does this:

1. Remove events older than window time interval with `ZREMRANGEBYSCORE`.
2. Count elements with `ZCARD`.
3. If the number of elements is less than limit, than add event (allow it). Otherwise - deny.

See full lua script in the [source code](https://github.com/tetafro/rate/blob/master/limiter.go#L109).

## Usage

```go
import "github.com/tetafro/rate"

// 1000 events per second
limit := NewLimit("localhost:6379", "rate-limiter", 1000)

for event := range events {
    if !limit.Allow() {
        continue // denied
    }
    doSomething() // allowed
}
```

## Testing

Run unit tests:
```sh
go test
```

Run integration tests (uses docker):
```sh
./integration_test
```
