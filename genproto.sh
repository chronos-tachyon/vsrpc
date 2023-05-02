#!/bin/bash
set -euo pipefail
umask 022
cd "$(dirname "$0")"
readonly module='module=github.com/chronos-tachyon/vsrpc'
find . -name "*.pb.go" -type f -delete
find proto -name "*.proto" -type f -print0 | xargs -0 protoc -Iproto --go_out="${module}:."
protoc -Iproto -Iexample --go_out="${module}:." --go-vsrpc_out="${module}:." example.proto
