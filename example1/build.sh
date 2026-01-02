#!/bin/sh
set -e

VERSION=$(git describe --tags --always)
REVISION=$(git rev-parse HEAD)
DATE=$(TZ=Asia/Seoul date +"%Y-%m-%d %H:%M %Z")

# 현재 폴더명 추출
APP=$(basename "$PWD")

go build -ldflags "-X 'main.Version=$VERSION' \
    -X 'main.Revision=$REVISION' \
    -X 'main.Date=$DATE'" -o "$APP"

./"$APP"
