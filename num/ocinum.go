// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package num

/*
#include <oci.h>
#include <stdlib.h>
*/
//import "C"
import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"
)

// OCINum is an OCINumber
//
// SQLT_VNU: 22 bytes, at max.
// SQLT_NUM: 21 bytes
//
// http://docs.oracle.com/cd/B28359_01/appdev.111/b28395/oci03typ.htm#sthref365
//
/*
Oracle stores values of the NUMBER datatype in a variable-length format. The first byte is the exponent and is followed by 1 to 20 mantissa bytes. The high-order bit of the exponent byte is the sign bit; it is set for positive numbers and it is cleared for negative numbers. The lower 7 bits represent the exponent, which is a base-100 digit with an offset of 65.

To calculate the decimal exponent, add 65 to the base-100 exponent and add another 128 if the number is positive. If the number is negative, you do the same, but subsequently the bits are inverted. For example, -5 has a base-100 exponent = 62 (0x3e). The decimal exponent is thus (~0x3e) -128 - 65 = 0xc1 -128 -65 = 193 -128 -65 = 0.

Each mantissa byte is a base-100 digit, in the range 1..100. For positive numbers, the digit has 1 added to it. So, the mantissa digit for the value 5 is 6. For negative numbers, instead of adding 1, the digit is subtracted from 101. So, the mantissa digit for the number -5 is 96 (101 - 5). Negative numbers have a byte containing 102 appended to the data bytes. However, negative numbers that have 20 mantissa bytes do not have the trailing 102 byte. Because the mantissa digits are stored in base 100, each byte can represent 2 decimal digits. The mantissa is normalized; leading zeroes are not stored.

Up to 20 data bytes can represent the mantissa. However, only 19 are guaranteed to be accurate. The 19 data bytes, each representing a base-100 digit, yield a maximum precision of 38 digits for an Oracle NUMBER.

If you specify the datatype code 2 in the dty parameter of an OCIDefineByPos() call, your program receives numeric data in this Oracle internal format. The output variable should be a 21-byte array to accommodate the largest possible number. Note that only the bytes that represent the number are returned. There is no blank padding or NULL termination. If you need to know the number of bytes returned, use the VARNUM external datatype instead of NUMBER.
*/
type OCINum []byte

// Print the number into the given byte slice.
func (num OCINum) Print(buf []byte) []byte {
	res := buf[:0]
	if bytes.Equal(num, []byte{128}) {
		return append(res, '0')
	}
	b := num[0]
	negative := b&(1<<7) == 0
	exp := int16(b & ((1 << 7) - 1))
	fmt.Printf("%v negative? %t; exp=%d\n", buf, negative, exp)
	if negative {
		res = append(res, '-')
		exp = ^exp - 128 - 65
		length := len(num)
		if num[length-1] == 102 {
			length--
		}
		for i, b := range num[1:length] {
			j := int64(101 - b)
			if i != 0 && j < 10 {
				res = append(res, '0')
			}
			res = strconv.AppendInt(res, j, 10)
		}
	} else {
		exp += 65 + 128
		for i, b := range num[1:] {
			fmt.Printf("i=%d\n", b-1)
			j := int64(b - 1)
			if i != 0 && j < 10 {
				res = append(res, '0')
			}
			res = strconv.AppendInt(res, j, 10)
		}
	}
	return res
}

var bytesPool = sync.Pool{New: func() interface{} { return make([]byte, 0, 4) }}

// String returns the string representation of the number.
func (num OCINum) String() string {
	b := bytesPool.Get().([]byte)
	s := string(num.Print(b))
	bytesPool.Put(b)
	return s
}

// SetString sets the OCINum to the number in s.
func (num *OCINum) SetString(s string) error {
	return errors.New("not implemented")
}
