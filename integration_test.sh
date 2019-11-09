#!/bin/bash

docker rm -f redis_rate_test 2>/dev/null
docker run -d --rm -p 127.0.0.1:6379:6379 --name redis_rate_test redis:3.2 > /dev/null
go test -tags=integration
docker stop redis_rate_test > /dev/null
