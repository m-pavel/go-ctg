## Golang bindings for the IBM CICS CTG

## Pre-requirements
  1. Download IBM CTG SDK package from ftp://public.dhe.ibm.com/software/htp/cics/support/supportpacs/individual. E.g. CICS_TG_SDK_92_Unix.tar.gz
  2. Unpack 
     - cicstgsdk/api/c/remote/include to ctg-api/include
     - cicstgsdk/api/c/remote/runtime/LinuxI/lib64 to ctg-api/lib64
  3. Usage 
  ```go
        ctg, err := Connect("host", 6363, 60)
    	if err != nil {
    		panic(err)
    	}
    	defer ctg.Close()
    
       	p := CTGClientParams{Server: "server", Transaction: "TRAN", 
                User: "user", Password: "password"}
    	request := makeRequestBytes()
        respBytes, err := ctg.Call(p, "PROGRAM", request)
    	if err != nil {
    		panic(err)
    	}
        processResponseBytes(respBytes)
```   
   