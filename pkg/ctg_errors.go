package pkg

// #include <ctgclient.h>
import "C"
import "fmt"

var ctgMessages = make(map[int]string)

func init() {
	ctgMessages[0] = "No error"
	ctgMessages[-1] = "Invalid data length"
	ctgMessages[-2] = "Invalid extend mode"
	ctgMessages[-3] = "No CICS"
	ctgMessages[-4] = "CICS died"
	ctgMessages[-5] = "Request timeout"
	ctgMessages[-6] = "Response timeout"
	ctgMessages[-7] = "Transaction abend"
	ctgMessages[-8] = "Error LUW token"
	ctgMessages[-9] = "system error"
	ctgMessages[-10] = "Null win handle"
	ctgMessages[-12] = "Null message id"
	ctgMessages[-19] = "Invalid data area"
	ctgMessages[-21] = "Invalid verion"
	ctgMessages[-14] = "Invalid call type"

}

func ctgError(res C.int) error {
	return ctgErrorAbend(res, [4]C.char{})
}

func ctgErrorAbend(res C.int, acode [4]C.char) error {
	if res == C.CTG_OK {
		return nil
	}
	desc, ok := ctgMessages[int(res)]
	if !ok {
		desc = "Unknown"
	}
	astring := make([]byte, 0)
	for i := 0; i < len(acode); i++ {
		if acode[i] == 0x00 {
			break
		}
		astring = append(astring, byte(acode[i]))
	}
	return fmt.Errorf("CTG error [%d] {%s}: %s", res, string(astring), desc)
}
