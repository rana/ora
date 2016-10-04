// +build ignore

// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package main

import (
	"bytes"
	"io/ioutil"
	"log"
)

func main() {
	src := "defFloat64.go"
	b := readFile(src)
	dst := "defFloat32.go"
	log.Printf("%s => %s", src, dst)
	if err := ioutil.WriteFile(
		dst,
		bytes.Replace(b, []byte("64"), []byte("32"), -1),
		0644,
	); err != nil {
		log.Fatal(err)
	}

	src = "defInt64.go"
	b = readFile(src)
	for _, s := range []string{"8", "16", "32"} {
		dst = "defInt" + s + ".go"
		log.Printf("%s => %s", src, dst)
		if err := ioutil.WriteFile(
			dst,
			bytes.Replace(b, []byte("64"), []byte(s), -1),
			0644,
		); err != nil {
			log.Fatal(err)
		}
	}

	for _, pair := range [][2]string{
		{"OCI_NUMBER_SIGNED", "OCI_NUMBER_UNSIGNED"},
		{"int64", "uint64"},
		{"Int64", "Uint64"},
	} {
		b = bytes.Replace(b, []byte(pair[0]), []byte(pair[1]), -1)
	}
	dst = "defUint64.go"
	if err := ioutil.WriteFile(dst, b, 0644); err != nil {
		log.Fatal(err)
	}
	for _, s := range []string{"8", "16", "32"} {
		dst = "defUint" + s + ".go"
		log.Printf("%s => %s", src, dst)
		if err := ioutil.WriteFile(
			dst,
			bytes.Replace(b, []byte("64"), []byte(s), -1),
			0644,
		); err != nil {
			log.Fatal(err)
		}
	}

	for _, plus := range []string{"", "Ptr", "Slice"} {
		src = "bndFloat64" + plus + ".go"
		b := readFile(src)
		dst = "bndFloat32" + plus + ".go"
		log.Printf("%s => %s", src, dst)
		if err := ioutil.WriteFile(
			dst,
			bytes.Replace(
				bytes.Replace(b, []byte("64"), []byte("32"), -1),
				[]byte("floatSixtyFour("), []byte("float64("), -1),
			0644,
		); err != nil {
			log.Fatal(err)
		}

		src = "bndInt64" + plus + ".go"
		b = readFile(src)
		for _, s := range []string{"8", "16", "32"} {
			dst = "bndInt" + s + plus + ".go"
			log.Printf("%s => %s", src, dst)
			if err := ioutil.WriteFile(
				dst,
				bytes.Replace(
					bytes.Replace(b, []byte("64"), []byte(s), -1),
					[]byte("intSixtyFour("), []byte("int64("), -1),
				0644,
			); err != nil {
				log.Fatal(err)
			}
		}

		for _, pair := range [][2]string{
			{"OCI_NUMBER_SIGNED", "OCI_NUMBER_UNSIGNED"},
			{"int64", "uint64"},
			{"Int64", "Uint64"},
		} {
			b = bytes.Replace(b, []byte(pair[0]), []byte(pair[1]), -1)
		}
		dst = "bndUint64" + plus + ".go"
		if err := ioutil.WriteFile(dst, b, 0644); err != nil {
			log.Fatal(err)
		}
		for _, s := range []string{"8", "16", "32"} {
			dst = "bndUint" + s + plus + ".go"
			log.Printf("%s => %s", src, dst)
			if err := ioutil.WriteFile(
				dst,
				bytes.Replace(
					bytes.Replace(b, []byte("64"), []byte(s), -1),
					[]byte("intSixtyFour("), []byte("int64("), -1),
				0644,
			); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func readFile(fn string) []byte {
	src, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	src = bytes.Replace(src, []byte("//go:generate "), []byte("// Generated from "+fn+" by "), -1)
	return src
}
