package cobol

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type CobolType byte

type cobolField struct {
	cname  string
	ctype  string
	csize  int
	casize int
	csign  bool
}

func cobolSize(i interface{}) (int, error) {
	size := 0
	err := iterateOverCobol(reflect.Indirect(reflect.ValueOf(i)), func(ct *cobolField, root reflect.Type, rv reflect.Value, f reflect.StructField, v reflect.Value) error {
		size += ct.csize
		return nil
	})
	return size, err
}

func cobolType(t string, f *reflect.StructField) (*cobolField, error) {
	rs := strings.Split(t, ",")
	cf := cobolField{}
	ctyped := t
	if len(rs) == 2 {
		cf.cname = rs[0]
		ctyped = rs[1]
	} else {
		cf.cname = strings.ToUpper(f.Name)
	}
	if ctyped[0:1] == "S" {
		cf.csign = true
		ctyped = ctyped[1:]
	}

	rxp := regexp.MustCompile(`(\w+)\((\d+)\)`)
	rrs := rxp.FindStringSubmatch(ctyped)
	if len(rrs) != 0 {
		cf.ctype = rrs[1]
		v, err := strconv.Atoi(rrs[2])
		if err != nil {
			return nil, fmt.Errorf("Wrong cobol type tag %s", t)
		}
		cf.csize = v
	} else {
		cf.ctype = ctyped
		cf.csize = 1
	}
	cf.casize = cf.csize

	if strings.Index(ctyped, "COMP-3") > 0 {
		cf.ctype = "COMP-3"
		if cf.csize%2 == 0 {
			cf.csize += 2
		} else {
			cf.csize += 1
		}
		cf.csize = cf.csize / 2
	} else {
		if strings.Index(ctyped, "COMP") > 0 {
			cf.ctype = "COMP"
			if cf.csize < 5 {
				cf.csize = 2
			} else if cf.csize > 4 && cf.csize < 10 {
				cf.csize = 4
			} else if cf.csize > 9 {
				cf.csize = 8
			}
		} else {
			if cf.csign {
				cf.csize++
			}
		}
	}
	return &cf, nil
}

func iterateOverCobol(v reflect.Value, cbk func(ct *cobolField, root reflect.Type, rv reflect.Value, f reflect.StructField, v reflect.Value) error) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("cobol")
		if tag == "" || tag == "-" {
			if v.Field(i).Kind() == reflect.Ptr || v.Field(i).Kind() == reflect.Struct {
				tb := v.Field(i)
				if err := iterateOverCobol(tb, cbk); err != nil {
					return err
				}
			} else {
				log.Printf("No tag on %s\n", t.Field(i).Name)
			}
			continue
		}
		f := t.Field(i)
		ct, err := cobolType(tag, &f)
		if err != nil {
			return err
		}
		if err := cbk(ct, t, v, f, v.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

func LogCobol(ci interface{}) {
	err := iterateOverCobol(reflect.Indirect(reflect.ValueOf(ci)), func(ct *cobolField, root reflect.Type, rv reflect.Value, f reflect.StructField, v reflect.Value) error {
		fmt.Fprintf(os.Stdout, "%16s : %-v\n", ct.cname, reflect.Indirect(v).Interface())
		return nil
	})
	if err != nil {
		panic(err)
	}
}
