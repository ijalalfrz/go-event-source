#!/bin/sh

# have to use postgres (super user) because test-fixtures need to disable system triggers
# when loading test fixtures
# use test DB
export DB_DSN=postgres://postgres@postgres/transaction_test?sslmode=disable
export LOG_LEVEL=debug
export GOCOVERDIR=test/coverage

cd /app

# build & run server
go build -cover -mod=vendor -o bin/app cmd/main.go
bin/app http &

# save last bg process' PID so we can kill it later
SERVER_PID=$!
echo "server pid: $SERVER_PID"

# run cucumber
go test --tags=integration -v -count=1 ./...
exitcode=$?

# kill server
kill $SERVER_PID

# print coverage
go tool covdata percent -i=$GOCOVERDIR
output=coverage.out
go tool covdata textfmt -i=$GOCOVERDIR -o $output
go tool cover -func $output | grep total:

exit $exitcode
