package ora

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// tags `db:column_name,pk,fk1,fk2,fk3,fk4,omit`

var ormTblNames map[string]string = make(map[string]string)
var ormTblCols map[string][]col = make(map[string][]col)

// Represents a result type returned by the ora.Sel method.
type ResType int

const (
	// Represents a slice of pointers returned by the ora.Sel method.
	// The value type is a struct determined by the user.
	SliceOfPtr ResType = iota
	// Represents a slice of values returned by the ora.Sel method.
	// The value type is a struct determined by the user.
	SliceOfVal
	// Represents a map of pointers returned by the ora.Sel method.
	// The value type is a struct determined by the user.
	MapOfPkPtr
	MapOfFk1Ptr
	MapOfFk2Ptr
	MapOfFk3Ptr
	MapOfFk4Ptr
	// Represents a map of values returned by the ora.Sel method.
	// The value type is a struct determined by the user.
	MapOfPkVal
	MapOfFk1Val
	MapOfFk2Val
	MapOfFk3Val
	MapOfFk4Val
)

// Represents attributes marked in the `db` StructTag.
type dbTag int

const (
	pk  dbTag = 1 << iota
	fk1 dbTag = 1 << iota
	fk2 dbTag = 1 << iota
	fk3 dbTag = 1 << iota
	fk4 dbTag = 1 << iota
)

type col struct {
	fieldIdx int
	name     string
	gct      GoColumnType
	attr     dbTag
}

func Ins(v interface{}, ses *Ses) (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errors.New(fmt.Sprint(value))
		}
	}()
	tblName, _, cols, err := getOrmCols(v)
	if err != nil {
		return err
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	pairs := make([]interface{}, len(cols)*2)
	for n, col := range cols {
		p := n * 2
		pairs[p] = col.name
		fv := rv.Field(col.fieldIdx)
		if n == len(cols)-1 {
			// ensure last field is pointer to capture db pk value
			if fv.Kind() == reflect.Ptr {
				pairs[p+1] = fv.Interface()
			} else {
				pairs[p+1] = fv.Addr().Interface()
			}
		} else {
			pairs[p+1] = fv.Interface()
		}
	}
	return ses.Ins(tblName, pairs...)
}

func Upd(v interface{}, ses *Ses) (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errors.New(fmt.Sprint(value))
		}
	}()
	tblName, _, cols, err := getOrmCols(v)
	if err != nil {
		return err
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	pairs := make([]interface{}, len(cols)*2)
	for n, col := range cols {
		p := n * 2
		pairs[p] = col.name
		pairs[p+1] = rv.Field(col.fieldIdx).Interface()
	}
	return ses.Upd(tblName, pairs...)
}

func Del(v interface{}, ses *Ses) (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errors.New(fmt.Sprint(value))
		}
	}()
	tblName, _, cols, err := getOrmCols(v)
	if err != nil {
		return err
	}
	lastCol := cols[len(cols)-1]
	var buf bytes.Buffer
	buf.WriteString("DELETE FROM ")
	buf.WriteString(tblName)
	buf.WriteString(" WHERE ")
	buf.WriteString(lastCol.name)
	buf.WriteString(" = :VAL")
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	_, err = ses.PrepAndExe(buf.String(), rv.Field(lastCol.fieldIdx).Interface())
	return err
}

func Sel(v interface{}, rt ResType, ses *Ses, where string, whereArgs ...interface{}) (result interface{}, err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errors.New(fmt.Sprint(value))
		}
	}()
	tblName, t, cols, err := getOrmCols(v)
	if err != nil {
		return nil, err
	}
	// build SELECT statement, GoColumnTypes
	gcts := make([]GoColumnType, len(cols))
	buf := new(bytes.Buffer)
	buf.WriteString("SELECT ")
	for n, col := range cols {
		buf.WriteString(col.name)
		if n != len(cols)-1 {
			buf.WriteString(", ")
		}
		gcts[n] = col.gct
	}
	buf.WriteString(" FROM ")
	buf.WriteString(tblName)
	if where != "" {
		buf.WriteString(" ")
		whereIdx := strings.Index(strings.ToUpper(where), "WHERE")
		if whereIdx < 0 {
			buf.WriteString("WHERE ")
		}
		buf.WriteString(where)
	}
	// prep
	stmt, err := ses.Prep(buf.String(), gcts...)
	if err != nil {
		defer stmt.Close()
		return nil, err
	}
	// qry
	rset, err := stmt.Qry(whereArgs...)
	if err != nil {
		defer stmt.Close()
		return nil, err
	}
	switch rt {
	case SliceOfPtr:
		sliceT := reflect.SliceOf(reflect.New(t).Type())
		sliceOfPtrRV := reflect.MakeSlice(sliceT, 0, 0)
		for rset.Next() {
			ptrRV := reflect.New(t)
			valRV := ptrRV.Elem()
			for n := range cols {
				f := valRV.Field(cols[n].fieldIdx)
				f.Set(reflect.ValueOf(rset.Row[n]))
			}
			sliceOfPtrRV = reflect.Append(sliceOfPtrRV, ptrRV)
		}
		result = sliceOfPtrRV.Interface()
	case SliceOfVal:
		sliceT := reflect.SliceOf(t)
		sliceOfValRV := reflect.MakeSlice(sliceT, 0, 0)
		for rset.Next() {
			valRV := reflect.New(t).Elem()
			for n := range cols {
				f := valRV.Field(cols[n].fieldIdx)
				f.Set(reflect.ValueOf(rset.Row[n]))
			}
			sliceOfValRV = reflect.Append(sliceOfValRV, valRV)
		}
		result = sliceOfValRV.Interface()
	case MapOfPkPtr, MapOfFk1Ptr, MapOfFk2Ptr, MapOfFk3Ptr, MapOfFk4Ptr:
		// lookup column for map key
		var keyRT reflect.Type
		switch rt {
		case MapOfPkPtr:
			for _, col := range cols {
				if col.attr&pk != 0 {
					keyRT = t.Field(col.fieldIdx).Type
					break
				}
			}
			if keyRT == nil {
				return nil, errors.New(fmt.Sprintf("Unable to make a map of pk to pointers for struct '%v'. '%v' doesn't have an exported field marked with a `db:\",pk\"` tag.", t.Name(), t.Name()))
			}
		case MapOfFk1Ptr:
			for _, col := range cols {
				if col.attr&fk1 != 0 {
					keyRT = t.Field(col.fieldIdx).Type
					break
				}
			}
			if keyRT == nil {
				return nil, errors.New(fmt.Sprintf("Unable to make a map of fk1 to pointers for struct '%v'. '%v' doesn't have an exported field marked with a `db:\",fk1\"` tag.", t.Name(), t.Name()))
			}
		case MapOfFk2Ptr:
			for _, col := range cols {
				if col.attr&fk2 != 0 {
					keyRT = t.Field(col.fieldIdx).Type
					break
				}
			}
			if keyRT == nil {
				return nil, errors.New(fmt.Sprintf("Unable to make a map of fk2 to pointers for struct '%v'. '%v' doesn't have an exported field marked with a `db:\",fk2\"` tag.", t.Name(), t.Name()))
			}
		case MapOfFk3Ptr:
			for _, col := range cols {
				if col.attr&fk3 != 0 {
					keyRT = t.Field(col.fieldIdx).Type
					break
				}
			}
			if keyRT == nil {
				return nil, errors.New(fmt.Sprintf("Unable to make a map of fk3 to pointers for struct '%v'. '%v' doesn't have an exported field marked with a `db:\",fk3\"` tag.", t.Name(), t.Name()))
			}
		case MapOfFk4Ptr:
			for _, col := range cols {
				if col.attr&fk4 != 0 {
					keyRT = t.Field(col.fieldIdx).Type
					break
				}
			}
			if keyRT == nil {
				return nil, errors.New(fmt.Sprintf("Unable to make a map of fk4 to pointers for struct '%v'. '%v' doesn't have an exported field marked with a `db:\",fk4\"` tag.", t.Name(), t.Name()))
			}
		}
		mapT := reflect.MapOf(keyRT, reflect.New(t).Type())
		mapOfPtrRV := reflect.MakeMap(mapT)
		for rset.Next() {
			var keyRV reflect.Value
			ptrRV := reflect.New(t)
			valRV := ptrRV.Elem()
			for n, col := range cols {
				f := valRV.Field(cols[n].fieldIdx)
				fv := reflect.ValueOf(rset.Row[n])
				f.Set(fv)
				switch rt {
				case MapOfPkPtr:
					if col.attr&pk != 0 { // validation ensures only one field is marked with `pk`
						keyRV = fv
					}
				case MapOfFk1Ptr:
					if col.attr&fk1 != 0 { // validation ensures only one field is marked with `fk1`
						keyRV = fv
					}
				case MapOfFk2Ptr:
					if col.attr&fk2 != 0 { // validation ensures only one field is marked with `fk2`
						keyRV = fv
					}
				case MapOfFk3Ptr:
					if col.attr&fk3 != 0 { // validation ensures only one field is marked with `fk3`
						keyRV = fv
					}
				case MapOfFk4Ptr:
					if col.attr&fk4 != 0 { // validation ensures only one field is marked with `fk4`
						keyRV = fv
					}
				}
			}
			mapOfPtrRV.SetMapIndex(keyRV, ptrRV)
		}
		result = mapOfPtrRV.Interface()
	case MapOfPkVal:
		// lookup column for map key
		var keyRT reflect.Type
		switch rt {
		case MapOfPkPtr:
			for _, col := range cols {
				if col.attr&pk != 0 {
					keyRT = t.Field(col.fieldIdx).Type
					break
				}
			}
			if keyRT == nil {
				return nil, errors.New(fmt.Sprintf("Unable to make a map of pk to values for struct '%v'. '%v' doesn't have an exported field marked with a `db:\",pk\"` tag.", t.Name(), t.Name()))
			}
		case MapOfFk1Ptr:
			for _, col := range cols {
				if col.attr&fk1 != 0 {
					keyRT = t.Field(col.fieldIdx).Type
					break
				}
			}
			if keyRT == nil {
				return nil, errors.New(fmt.Sprintf("Unable to make a map of fk1 to values for struct '%v'. '%v' doesn't have an exported field marked with a `db:\",fk1\"` tag.", t.Name(), t.Name()))
			}
		case MapOfFk2Ptr:
			for _, col := range cols {
				if col.attr&fk2 != 0 {
					keyRT = t.Field(col.fieldIdx).Type
					break
				}
			}
			if keyRT == nil {
				return nil, errors.New(fmt.Sprintf("Unable to make a map of fk2 to values for struct '%v'. '%v' doesn't have an exported field marked with a `db:\",fk2\"` tag.", t.Name(), t.Name()))
			}
		case MapOfFk3Ptr:
			for _, col := range cols {
				if col.attr&fk3 != 0 {
					keyRT = t.Field(col.fieldIdx).Type
					break
				}
			}
			if keyRT == nil {
				return nil, errors.New(fmt.Sprintf("Unable to make a map of fk3 to values for struct '%v'. '%v' doesn't have an exported field marked with a `db:\",fk3\"` tag.", t.Name(), t.Name()))
			}
		case MapOfFk4Ptr:
			for _, col := range cols {
				if col.attr&fk4 != 0 {
					keyRT = t.Field(col.fieldIdx).Type
					break
				}
			}
			if keyRT == nil {
				return nil, errors.New(fmt.Sprintf("Unable to make a map of fk4 to values for struct '%v'. '%v' doesn't have an exported field marked with a `db:\",fk4\"` tag.", t.Name(), t.Name()))
			}
		}
		mapT := reflect.MapOf(keyRT, t)
		mapOfValRV := reflect.MakeMap(mapT)
		for rset.Next() {
			var keyRV reflect.Value
			valRV := reflect.New(t).Elem()
			for n, col := range cols {
				f := valRV.Field(cols[n].fieldIdx)
				fv := reflect.ValueOf(rset.Row[n])
				f.Set(fv)
				switch rt {
				case MapOfPkPtr:
					if col.attr&pk != 0 { // validation ensures only one field is marked with `pk`
						keyRV = fv
					}
				case MapOfFk1Ptr:
					if col.attr&fk1 != 0 { // validation ensures only one field is marked with `fk1`
						keyRV = fv
					}
				case MapOfFk2Ptr:
					if col.attr&fk2 != 0 { // validation ensures only one field is marked with `fk2`
						keyRV = fv
					}
				case MapOfFk3Ptr:
					if col.attr&fk3 != 0 { // validation ensures only one field is marked with `fk3`
						keyRV = fv
					}
				case MapOfFk4Ptr:
					if col.attr&fk4 != 0 { // validation ensures only one field is marked with `fk4`
						keyRV = fv
					}
				}
			}
			mapOfValRV.SetMapIndex(keyRV, valRV)
		}
		result = mapOfValRV.Interface()
	}
	return result, err
}

func AddTable(v interface{}, tblName string) error {
	if v == nil {
		return errors.New("Unable to determine type from nil value.")
	}
	t := reflect.TypeOf(v)
	// set user-specified table name
	ormTblNames[strings.ToUpper(t.Name())] = strings.ToUpper(tblName)
	getOrmCols(v)
	return nil
}

func getOrmCols(v interface{}) (tblName string, t reflect.Type, cols []col, err error) {
	if v == nil {
		return "", nil, nil, errors.New("Unable to determine type from nil value.")
	}
	// get struct value
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	t = rv.Type()
	// lookup table name
	typeName := strings.ToUpper(t.Name())
	tblName = ormTblNames[typeName]
	if tblName == "" { // possible user passed in empty string for table name
		tblName = typeName
		ormTblNames[typeName] = tblName
	}
	// lookup cols
	cols, ok := ormTblCols[tblName]
	if !ok { // determine columns
		if t.Kind() == reflect.Struct {
			cols = make([]col, 0)
		Outer:
			for n := 0; n < t.NumField(); n++ {
				f := t.Field(n)
				// skip unexported fields
				if unicode.IsLower(rune(f.Name[0])) {
					continue
				}
				tag := f.Tag.Get("db")
				col := col{fieldIdx: n}
				if tag == "" { // no db tag; use field name
					col.name = f.Name
				} else {
					tagValues := strings.Split(tag, ",")
					for n := range tagValues {
						tagValues[n] = strings.ToLower(strings.Trim(tagValues[n], " "))
					}
					// check for `omit` field
					for _, tagValue := range tagValues {
						if tagValue == "omit" {
							continue Outer
						}
					}
					if len(tagValues) == 0 {
						return "", nil, nil, errors.New(fmt.Sprintf("Struct '%v' field '%v' has `db` tag but no value.", t.Name(), f.Name))
					} else {
						if tagValues[0] == "" { // may be empty string in case of `db:",pk"`
							col.name = f.Name
						} else {
							col.name = tagValues[0]
						}
						// check for single `pk`,`fk1`,`fk2`,`fk3`,`fk4` field
						pkCount := 0
						fk1Count := 0
						fk2Count := 0
						fk3Count := 0
						fk4Count := 0
						for _, tagValue := range tagValues {
							if tagValue == "pk" {
								col.attr |= pk
								pkCount++
							} else if tagValue == "fk1" {
								col.attr |= fk1
								fk1Count++
							} else if tagValue == "fk2" {
								col.attr |= fk2
								fk2Count++
							} else if tagValue == "fk3" {
								col.attr |= fk3
								fk3Count++
							} else if tagValue == "fk4" {
								col.attr |= fk4
								fk4Count++
							}
						}
						if pkCount > 1 {
							return "", nil, nil, errors.New(fmt.Sprintf("Struct '%v' has more than one exported field marked with a `db:\",pk\"` tag.", t.Name()))
						} else if fk1Count > 1 {
							return "", nil, nil, errors.New(fmt.Sprintf("Struct '%v' has more than one exported field marked with a `db:\",fk1\"` tag.", t.Name()))
						} else if fk2Count > 1 {
							return "", nil, nil, errors.New(fmt.Sprintf("Struct '%v' has more than one exported field marked with a `db:\",fk2\"` tag.", t.Name()))
						} else if fk3Count > 1 {
							return "", nil, nil, errors.New(fmt.Sprintf("Struct '%v' has more than one exported field marked with a `db:\",fk3\"` tag.", t.Name()))
						} else if fk4Count > 1 {
							return "", nil, nil, errors.New(fmt.Sprintf("Struct '%v' has more than one exported field marked with a `db:\",fk4\"` tag.", t.Name()))
						}
					}
				}
				col.name = strings.ToUpper(col.name)
				col.gct = gct(f.Type)
				cols = append(cols, col)
			}
			ormTblCols[tblName] = cols
			// place pk at last index for Ins, Upd
			for n, col := range cols {
				if col.attr&pk != 0 {
					cols[n], cols[len(cols)-1] = cols[len(cols)-1], cols[n]
				}
			}
		}
	}
	if len(cols) == 0 {
		return "", nil, nil, errors.New(fmt.Sprintf("Struct '%v' has no db columns.", t.Name()))
	}
	return tblName, t, cols, nil
}

func gct(rt reflect.Type) GoColumnType {
	switch rt.Kind() {
	case reflect.Bool:
		return B
	case reflect.String:
		return S
	case reflect.Array, reflect.Slice:
		name := rt.Elem().Name()
		if name == "uint8" || name == "byte" {
			return Bin
		}
	case reflect.Int64:
		return I64
	case reflect.Int32:
		return I32
	case reflect.Int16:
		return I16
	case reflect.Int8:
		return I8
	case reflect.Uint64:
		return U64
	case reflect.Uint32:
		return U32
	case reflect.Uint16:
		return U16
	case reflect.Uint8:
		return U8
	case reflect.Float64:
		return F64
	case reflect.Float32:
		return F32
	case reflect.Struct:
		name := rt.Name()
		switch rt.PkgPath() {
		case "time":
			if name == "Time" {
				return T
			}
		case "ora":
			switch name {
			case "OraI64":
				return OraI64
			case "OraI32":
				return OraI32
			case "OraI16":
				return OraI16
			case "OraI8":
				return OraI8
			case "OraU64":
				return OraU64
			case "OraU32":
				return OraU32
			case "OraU16":
				return OraU16
			case "OraU8":
				return OraU8
			case "OraF64":
				return OraF64
			case "OraF32":
				return OraF32
			case "OraT":
				return OraT
			case "OraS":
				return OraS
			case "OraB":
				return OraB
			case "OraBin":
				return OraBin
			}
		}
	}
	return D
}
