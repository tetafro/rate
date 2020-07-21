.PHONY: test
test: test-unit test-integration

.PHONY: test-unit
test-unit:
	go test -v -race
.PHONY: test-integration
test-integration:
	go test -v -race -tags=integration -run Integration

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=./profile.out -covermode=atomic
	go tool cover -html=./profile.out
