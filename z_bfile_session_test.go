// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"testing"

	"gopkg.in/rana/ora.v4"
)

//// bfile
//bfile     oracleColumnType = "bfile not null"
//bfileNull oracleColumnType = "bfile null"

////////////////////////////////////////////////////////////////////////////////
// bfile
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_bfile(t *testing.T) {
	sc := ora.NewStmtCfg()
	t.Run("bfile", func(t *testing.T) {
		t.Parallel()
		//enableLogging(t)
		testBindDefine(gen_OraBfile(false), bfile, t, sc)
	})
	for _, isNull := range []bool{true, false} {
		t.Run(fmt.Sprintf("bfileNull_%t", isNull), func(t *testing.T) {
			t.Parallel()
			//enableLogging(t)
			testBindDefine(gen_OraBfile(isNull), bfileNull, t, sc)
		})
	}
}
