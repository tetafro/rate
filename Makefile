.PHONY: test
test:
	go test -tags=integration \
		-coverprofile=./profile.out \
		-covermode=atomic

.PHONY: test-unit
test-unit:
	go test -v -race

.PHONY: test-integration
test-integration:
	go test -v -race -tags=integration -run Integration

.PHONY: test-coverage
test-coverage: test
	go tool cover -html=./profile.out

.PHONY: lint
lint:
	golangci-lint run
