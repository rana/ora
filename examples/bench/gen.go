package main

//go:generate mkdir -p ora
//go:generate sh -c "sed -e 's:github.com/mattn/go-oci8:gopkg.in/rana/ora.v4:;s/Oci8/Ora/g;s/oci8/ora/g' oci8/bench_oci8.go >ora/bench_ora.go"
