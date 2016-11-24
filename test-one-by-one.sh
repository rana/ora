#!/bin/sh
n="${1:-10}"

go install -race && go test -race -c && \
grep -h '^func Test' z_*.go|cut -c10-|cut -d'(' -f1 \
| xargs -n "$n" | sed -e 's/ /|/g' | while read nm;
do
	echo "$nm"
	./ora.v4.test -test.v -test.run=$nm || break
done

