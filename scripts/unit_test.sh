#!/bin/sh

output=coverage.out
go test --tags=unit -v -timeout 10s -count=1 -cover ./... -args -test.gocoverdir="$PWD/test/coverage-unit"
exitcode=$?

# print coverage
go tool covdata percent -i=test/coverage-unit
output=coverage-unit.out
go tool covdata textfmt -i=test/coverage-unit -o $output
go tool cover -func $output | grep total:


exit $exitcode
