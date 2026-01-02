#!/bin/sh
set -e  # 에러 발생 시 즉시 중지

go build .
./example
