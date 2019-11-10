#!/bin/bash

docker rm -f redis_rate_test 1>/dev/null 2>/dev/null
docker run -d --rm -p 127.0.0.1:6379:6379 --name redis_rate_test redis:3.2 > /dev/null
go test -v -race \
    -tags=integration \
    -coverprofile=./profile.out \
    -covermode=atomic
docker stop redis_rate_test > /dev/null
