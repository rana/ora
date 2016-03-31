#!/bin/sh -e
DSN="$1"
BENCHTIME="${BENCHTIME:-1s}"
N=${N:-5000}
if [ -z "$DSN" ]; then
	DSN=${GO_ORA_DRV_TEST_USERNAME}/${GO_ORA_DRV_TEST_PASSWORD}@${GO_ORA_DRV_TEST_DB}
fi
echo "Compile ..." >&2
go build -o oci8.bench ./oci8 &
go build -o ora.bench ./ora
wait

run () {
	nm="$1"
	echo "$nm" >&2
	./"$nm".bench -cpuprofile="$nm".pprof -N=$N -test.benchtime=$BENCHTIME "$DSN"
	echo "cum\ntop20" | go tool pprof ./"$nm".bench "$nm".pprof
}

echo "Benchmarks on DSN=${DSN} ..." >&2
run oci8
run ora
