#!/bin/sh -e
DSN="$1"
BENCHTIME="${BENCHTIME:-1s}"
ROWCOUNT=5000
if [ -z "$DSN" ]; then
	DSN=${GO_ORA_DRV_TEST_USERNAME}/${GO_ORA_DRV_TEST_PASSWORD}@${GO_ORA_DRV_TEST_DB}
fi
echo "Compile ..." >&2
go build -o oci8.bench ./oci8 &
go build -o ora.bench ./ora
wait

echo "Benchmarks on DSN=${DSN} ..." >&2
echo "go-oci8" >&2
./oci8.bench -cpuprofile=oci8.pprof -N=$ROWCOUNT -test.benchtime=${BENCHTIME} "$DSN"
echo "cum\ntop20" | go tool pprof ./oci8.bench oci8.pprof

echo "ora" >&2
./ora.bench -cpuprofile=ora.pprof -N=$ROWCOUNT -test.benchtime=${BENCHTIME} "$DSN"
echo "cum\ntop20" | go tool pprof ./ora.bench ora.pprof
