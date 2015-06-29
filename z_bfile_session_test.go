// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"testing"
)

//// bfile
//bfile     oracleColumnType = "bfile not null"
//bfileNull oracleColumnType = "bfile null"

////////////////////////////////////////////////////////////////////////////////
// bfile
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_OraBfile_bfile_session(t *testing.T) {
	//enableLogging(t)
	testBindDefine(gen_OraBfile(false), bfile, t, nil)
}

func TestBindDefine_OraBfileNull_bfile_session(t *testing.T) {
	//enableLogging(t)
	testBindDefine(gen_OraBfile(false), bfileNull, t, nil)
}

func TestBindDefine_OraBfileNull_bfile_null_session(t *testing.T) {
	//enableLogging(t)
	testBindDefine(gen_OraBfile(true), bfileNull, t, nil)
}
