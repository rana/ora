#!/bin/sh
n="${1:-10}"

go install -race && go test -race -c && \
grep -h '^func Test' *_test.go|cut -c10-|cut -d'(' -f1 \
| sort -R | xargs -n "$n" | sed -e 's/ /|/g' | \
while read nm;
do
	echo "$nm"
	./ora.v4.test -test.run=$nm || break
done

