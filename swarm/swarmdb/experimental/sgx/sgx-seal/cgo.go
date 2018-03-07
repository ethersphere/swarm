
package main

/*
#include "./App/TEE.h"
#cgo LDFLAGS: -I./App -L. -ltee 
*/
import "C"

import (
	"fmt"
)

//export test
func test() {
    fmt.Printf("intel sgx hello\n")
}

func main() {
    C.sgx_seal()
    //test()
}
