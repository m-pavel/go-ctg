package ctg

// #cgo CFLAGS: -g -Wall -I../ctg-api/include -DCICS_LNX
// #cgo LDFLAGS: -L../ctg-api/lib64 -lctgclient
// #include <ctgclient.h>
// #include <ctgclient_eci.h>
import "C"
import (
	"log"
	"reflect"
	"unsafe"
)

type CTGClient struct {
	connToken C.CTG_ConnToken_t
}

type CTGClientParams struct {
	Server       string
	User         string
	Password     string
	Transaction  string
	Tpn          string
	CommareaSize int16
	Debug        bool
}

const DefaultCommAreaSize = 32500 // ctgclient_eci.h

func Connect(host string, port int, timeout int) (*CTGClient, error) {
	chost := C.CString(host)
	defer C.free(unsafe.Pointer(chost))
	var token C.CTG_ConnToken_t
	cli := CTGClient{connToken: token}
	res := C.CTG_openRemoteGatewayConnection(chost, C.int(port), &cli.connToken, C.int(timeout))
	return &cli, ctgError(res)
}

func (c *CTGClient) Call(params CTGClientParams, program string, commarea []byte) ([]byte, error) {
	eciParms := C.CTG_ECI_PARMS{}
	eciParms.eci_version = C.ECI_VERSION_2A
	eciParms.eci_call_type = C.ECI_SYNC

	commAreaSize := params.CommareaSize
	if commAreaSize == 0 {
		commAreaSize = DefaultCommAreaSize
	}

	creq := make([]byte, commAreaSize)
	copy(creq, commarea)
	sreq := C.CString(string(creq))
	eciParms.eci_commarea = unsafe.Pointer(sreq)
	eciParms.eci_commarea_length = C.short(commAreaSize)
	eciParms.commarea_outbound_length = C.short(len(commarea))
	eciParms.commarea_inbound_length = 0

	eciParms.eci_extend_mode = C.ECI_NO_EXTEND
	eciParms.eci_luw_token = C.ECI_LUW_NEW
	eciParms.eci_timeout = 0

	rarr := [8]C.char{}
	copy(rarr[:], charArr(params.Server))
	eciParms.eci_system_name = rarr

	parr := [8]C.char{}
	copy(parr[:], charArr(program))
	eciParms.eci_program_name = parr

	tarr := [4]C.char{}
	copy(tarr[:], charArr(params.Transaction))
	eciParms.eci_transid = tarr

	tprarr := [4]C.char{}
	copy(tprarr[:], charArr(params.Tpn))
	eciParms.eci_tpn = tprarr

	eciParms.eci_userid_ptr = C.CString(params.User)
	eciParms.eci_password_ptr = C.CString(params.Password)
	if params.Debug {
		log.Println(eciParms)
	}
	res := C.CTG_ECI_Execute(c.connToken, &eciParms)
	if params.Debug {
		log.Println(eciParms)
		log.Println(eciParms.eci_commarea_length)
		log.Println(eciParms.commarea_inbound_length)
	}
	// TODO check for leak
	var list []byte
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&list)))
	sliceHeader.Cap = int(commAreaSize)
	sliceHeader.Len = int(commAreaSize)
	sliceHeader.Data = uintptr(unsafe.Pointer(eciParms.eci_commarea))

	return list, ctgErrorAbend(res, eciParms.eci_abend_code)
}

func charArr(s string) []C.char {
	res := make([]C.char, len(s))
	for i := range s {
		res[i] = C.char(s[i])
	}
	return res
}

func (c *CTGClient) Close() error {
	log.Println("Disconnecting...")
	if c.connToken != nil {
		cc := c.connToken
		defer func() {
			c.connToken = nil
		}()
		return ctgError(C.CTG_closeGatewayConnection(&cc))
	}
	return nil
}
