#!/bin/sh
n="${1:-10}"

PRG=./ora.v4.test
if time echo '' >/dev/null 2>&1; then
	PRG="time $PRG"
elif /usr/bin/time echo '' >/dev/null 2>&1; then
	PRG="/usr/bin/time $PRG"
fi

go install -race && go test -race -c && \
grep -h '^func Test' *_test.go|cut -c10-|cut -d'(' -f1 \
| sort -R | xargs -n "$n" | sed -e 's/ /|/g' | \
while read nm;
do
	echo ''
	echo "$nm"
	$PRG -test.run=$nm || break
done

