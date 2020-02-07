package cobol

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

var asciiTab = []byte{0x00, 0x01, 0x02, 0x03, 0x04,
	0x09, 0x06, 0x7f, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x08, 0x17, 0x18, 0x19, 0x1a,
	0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22, 0x23, 0x24, 0x0a,
	0x17, 0x1b, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x05, 0x06, 0x07, 0x30,
	0x31, 0x16, 0x33, 0x34, 0x35, 0x36, 0x04, 0x38, 0x39, 0x3a, 0x3b,
	0x14, 0x15, 0x3e, 0x1a, 0x20, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46,
	0x47, 0x48, 0x49, 0x8e, 0x2e, 0x3c, 0x28, 0x2b, 0x4f, 0x26, 0x51,
	0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x9a, 0x24, 0x2a,
	0x29, 0x3b, 0x5e, 0x2d, 0x2f, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67,
	0x68, 0x69, 0x94, 0x2c, 0x25, 0x5f, 0x3e, 0x3f, 0x70, 0x71, 0x72,
	0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x60, 0x3a, 0x23, 0x40, 0x27,
	0x3d, 0x22, 0x80, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68,
	0x69, 0x8a, 0x7b, 0x8c, 0x8d, 0x8e, 0x8f, 0x90, 0x6a, 0x6b, 0x6c,
	0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x9a, 0x7d, 0x9c, 0x9d, 0x9e,
	0x9f, 0xa0, 0x7e, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a,
	0xaa, 0xab, 0xac, 0x5b, 0xae, 0xaf, 0xb0, 0xb1, 0xb2, 0xb3, 0xb4,
	0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xbb, 0xbc, 0x5d, 0xbe, 0xbf,
	0x84, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0xca,
	0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0x81, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e,
	0x4f, 0x50, 0x51, 0x52, 0xda, 0xdb, 0xdc, 0xdd, 0xde, 0xdf, 0x5c,
	0xe1, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0xea, 0xeb,
	0xec, 0xed, 0xee, 0xef, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36,
	0x37, 0x38, 0x39, 0x7c, 0xfb, 0xfc, 0xfd, 0xfe, 0xff}

type Decoder struct {
	Debug bool
}

func (d Decoder) Decode(cbl []byte, tov interface{}) error {
	bi := 0
	return iterateOverCobol(reflect.Indirect(reflect.ValueOf(tov)), func(ct *cobolField, rootType reflect.Type, rootValue reflect.Value, f reflect.StructField, v reflect.Value) error {
		if !v.CanSet() {
			return fmt.Errorf("Can't set %v", f)
		}
		fv, err := d.extractValue(ct, cbl, bi, rootType, rootValue, f, v)
		if err != nil {
			return err
		}
		if d.Debug {
			log.Printf("%s : %v\n", ct.cname, fv)
		}
		v.Set(fv)

		bi += ct.csize
		return nil
	})
}

func (d Decoder) extractValue(ct *cobolField, cbl []byte, bi int, rootType reflect.Type, rootValue reflect.Value, f reflect.StructField, v reflect.Value) (reflect.Value, error) {
	switch ct.ctype {
	case "X":
		bar := make([]byte, ct.csize)
		ebcToAsc(cbl[bi:bi+ct.csize], bar)
		switch v.Kind() {
		case reflect.String:
			return reflect.ValueOf(byteToString(bar)), nil
		case reflect.Ptr:
			str := byteToString(bar)
			return reflect.ValueOf(&str), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16:
			return reflect.ValueOf(bar[0]), nil
		default:
			tt := reflect.Indirect(reflect.ValueOf(bar)).Type()
			if !v.Type().AssignableTo(tt) {
				return reflect.ValueOf(nil), fmt.Errorf("%s field type is not assignable to %s", f.Name, tt.Name())
			}
			return reflect.ValueOf(bar), nil
		}
	case "S9", "9":
		iv, err := ebcToPic(cbl[bi : bi+ct.csize])
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		if v.Kind() == reflect.Ptr {
			return reflect.ValueOf(&iv), err
		}
		return reflect.ValueOf(iv), err
	case "COMP-3":
		iv, err := ebcComp3ToPic(cbl[bi:bi+ct.csize], ct.casize)
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		if v.Kind() == reflect.Ptr {
			return reflect.ValueOf(&iv), err
		}
		return reflect.ValueOf(iv), err
	case "COMP":
		iv, err := ebcCompToPic(cbl[bi : bi+ct.csize])
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		if v.Kind() == reflect.Ptr {
			return reflect.ValueOf(&iv), err
		}
		return reflect.ValueOf(iv), err
	default:
		if f.Type.Kind() == reflect.Array || f.Type.Kind() == reflect.Slice {
			if _, tf := rootType.FieldByName(ct.ctype); !tf {
				return reflect.ValueOf(nil), fmt.Errorf("array %s size field not found %s", f.Name, ct.ctype)
			}
			as := rootValue.FieldByName(ct.ctype).Int()
			selem := f.Type.Elem()
			sls := reflect.MakeSlice(reflect.SliceOf(selem), int(as), int(as))
			for i := 0; i < int(as); i++ {
				elem := reflect.New(selem)
				if esz, err := cobolSize(elem.Interface()); err != nil {
					return reflect.ValueOf(nil), err
				} else {
					if err := d.Decode(cbl[bi:bi+esz], elem.Interface()); err != nil {
						return reflect.ValueOf(nil), err
					}
					sls.Index(i).Set(elem.Elem())
					bi += esz
				}
			}
			//v.Set(sls)
			return sls, nil
		} else {
			return reflect.ValueOf(nil), fmt.Errorf("unsupported cobol type %s", ct.ctype)
		}
	}
}
func byteToString(b []byte) string {
	lc := len(b) - 1
	for ; lc >= 0; lc-- {
		if b[lc] != 0x00 {
			break
		}
	}
	return string(b[:lc+1])
}

func ebcToAsc(ebc []byte, ascii []byte) {
	//low := true
	space := true
	for i := len(ebc) - 1; i >= 0; i-- {
		ebcByte := ebc[i] & 0xff
		if ebcByte == 0x5a {
			ascii[i] = 0x21
		} else {
			ascii[i] = asciiTab[ebcByte]
		}
		//if ascii[i] != 0x00 {
		//	low = false
		//}
		if (ascii[i] == 0x00 || ascii[i] == 0x20) && space {
			ascii[i] = 0x00
		} else {
			space = false
		}
	}
}

func ebcToPic(ebc []byte) (int, error) {
	si := make([]byte, 0)
	for i := 0; i < len(ebc); i++ {
		ca := asciiTab[ebc[i]&0xff]
		if ca == 0x00 || ca == 0x20 {
			ca = '0'
		}
		si = append(si, ca)
	}
	return strconv.Atoi(string(si))
}

func ebcCompToPic(ebc []byte) (int64, error) {
	var res int64
	switch len(ebc) {
	case 2:
		res = int64((uint32(ebc[0])<<24)>>16 | (uint32(ebc[1])<<24)>>24)
	case 4:
		res = int64((uint32(ebc[0])<<24)>>0 | (uint32(ebc[1])<<24)>>8 | (uint32(ebc[2])<<24)>>16 | (uint32(ebc[3])<<24)>>24)
	case 8:
		res = int64(uint64(ebc[0])<<070&0xff00000000000000 | uint64(ebc[1])<<060&0xff000000000000 | uint64(ebc[2])<<050&0xff0000000000 |
			uint64(ebc[3])<<040&0xff00000000 | uint64(ebc[4])<<030&0xff000000 | uint64(ebc[5])<<020&0xff0000 | uint64(ebc[6])<<010&0xff00 | uint64(ebc[7])&0xff)
	}
	return res, nil
}

func ebcComp3ToPic(ebc []byte, aslen int) (int64, error) {
	as := make([]byte, 0)
	for i := 0; i < len(ebc)-1; i++ {
		as = append(as, (ebc[i]>>4)&0x0F+0x30, (ebc[i])&0x0F+0x30)
	}
	as = append(as, (ebc[len(ebc)-1]>>4)&0x0F+0x30)
	if ebc[len(ebc)-1]&0x0F == 0x0D {
		as = append([]byte{'-'}, as...)
	}
	res, err := strconv.Atoi(string(as))
	return int64(res), err
}
