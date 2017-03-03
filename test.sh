#!/bin/sh
set -e
set -x
ulimit -v unlimited
CGO_CFLAGS=-fsanitize=address CGO_LDFLAGS=-lasan go test -c
exec ./ora.v4.test "$@"
