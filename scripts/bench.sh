#!/usr/bin/env bash
set -euo pipefail
go test ./internal/server -run '^$' -bench 'BenchmarkPassthroughContract' -benchtime=1x -benchmem
