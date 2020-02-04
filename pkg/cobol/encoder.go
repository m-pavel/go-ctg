package cobol

import (
	"fmt"
	"reflect"
)

var ebcTab = []byte{0x00, 0x01, 0x02, 0x03, 0x37,
	0x2d, 0x2e, 0x2f, 0x16, 0x05, 0x25, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	0x10, 0x11, 0x12, 0x13, 0x3c, 0x3d, 0x32, 0x26, 0x18, 0x19, 0x3f,
	0x27, 0x1c, 0x1d, 0x1e, 0x1f, 0x40, 0x21, 0x7f, 0x7b, 0x5b, 0x6c,
	0x50, 0x7d, 0x4d, 0x5d, 0x5c, 0x4e, 0x6b, 0x60, 0x4b, 0x61, 0xf0,
	0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0x7a, 0x5e,
	0x4c, 0x7e, 0x6e, 0x6f, 0x7c, 0xc1, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6,
	0xc7, 0xc8, 0xc9, 0xd1, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8,
	0xd9, 0xe2, 0xe3, 0xe4, 0xe5, 0xe6, 0xe7, 0xe8, 0xe9, 0xad, 0xe0,
	0xbd, 0x5f, 0x6d, 0x79, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87,
	0x88, 0x89, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99,
	0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0x8b, 0xfa, 0x9b,
	0xa1, 0x07, 0x80, 0xd0, 0x82, 0x83, 0xc0, 0x85, 0x86, 0x87, 0x88,
	0x89, 0x8a, 0x8b, 0x8c, 0x8d, 0x4a, 0x8f, 0x90, 0x91, 0x92, 0x93,
	0x6a, 0x95, 0x96, 0x97, 0x98, 0xe0, 0x5a, 0x9b, 0x9c, 0x9d, 0x9e,
	0x9f, 0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9,
	0x5f, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb0, 0xb1, 0xb2, 0x4f, 0xb4,
	0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xbb, 0xbc, 0xbd, 0xbe, 0xbf,
	0xc0, 0xc1, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7, 0xc8, 0xc9, 0xca,
	0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0, 0xd1, 0xd2, 0xd3, 0xd4, 0xd5,
	0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xdb, 0xdc, 0xdd, 0xde, 0xdf, 0xe0,
	0xa1, 0xe2, 0xe3, 0xe4, 0xe5, 0xe6, 0xe7, 0xe8, 0xe9, 0xea, 0xeb,
	0xec, 0xed, 0xee, 0xef, 0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6,
	0xf7, 0xf8, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff}

type Encoder struct {
}

func (e Encoder) Encode(ic interface{}) ([]byte, error) {
	res := make([]byte, 0)
	err := iterateOverCobol(reflect.Indirect(reflect.ValueOf(ic)), func(ct *cobolField, root reflect.Type, rv reflect.Value, f reflect.StructField, v reflect.Value) error {
		switch ct.ctype {
		case "X":
			ebc, err := ascToEbc(v, ct.csize)
			if err != nil {
				return nil
			}
			res = append(res, ebc...)
		case "9":
			sv := itoa(v, ct.csize)
			ebc, err := ascToEbc(reflect.Indirect(reflect.ValueOf(sv)), ct.csize)
			if err != nil {
				return err
			}
			res = append(res, ebc...)
		case "COMP-3":
			res = append(res, comp3ToEbc(v, ct.csize)...)
		case "COMP":
			res = append(res, compToEbc(v, ct.csize)...)
		default:
			return fmt.Errorf("unsupported field type %s", ct.ctype)
		}
		//		log.Printf("%s %d %v\n", ct.cname, len(res), res)
		return nil
	})
	return res, err
}

func compToEbc(v reflect.Value, size int) []byte {
	val := v.Uint()
	res := make([]byte, size)
	for i := size - 1; i >= 0; i-- {
		bi := (size - i - 1) * 8
		res[i] = byte(val & (0xff << bi) >> bi)
	}
	return res
}
func comp3ToEbc(v reflect.Value, size int) []byte {
	res := make([]byte, size)

	sv := fmt.Sprintf("%d", v.Int())
	spos := len(sv)
	lo := uint8(0x0C)
	i := size
	for i > 0 && spos > 0 {
		hi := (sv[spos-1] - 0x30) << 4 & 0xF0
		spos--
		i--
		if hi != 0 || lo != 0 {
			res[i] = hi | lo
		}
		if spos != 0 {
			lo = (sv[spos-1] - 0x30) & 0x0F
			spos--
		} else {
			lo = 0
		}
		if lo != 0 {
			if i-1 > 0 {
				res[i-1] = lo
			} else {
				res[0] = lo
			}
		}
	}
	return res
}

func itoa(v reflect.Value, size int) []byte {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return make([]byte, size)
		} else {
			v = v.Elem()
		}
	}
	return []byte(fmt.Sprintf(fmt.Sprintf("%%0%dd", size), v.Int()))
}

func ascToEbc(v reflect.Value, size int) ([]byte, error) {
	if v.Kind() == reflect.Uint8 {
		return []byte{ebcTab[v.Uint()]}, nil
	}
	res := make([]byte, size)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return res, nil
	}
	if v.Len() > size {
		return nil, fmt.Errorf("Too big %v", v.Interface())
	}
	for i := 0; i < v.Len(); i++ {
		res[i] = ebcTab[v.Index(i).Uint()]
	}
	for i := v.Len(); i < size; i++ {
		res[i] = 64
	}

	return res, nil
}