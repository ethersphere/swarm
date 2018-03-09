
package main

/*
#include "../go-with-intel-sgx/App/TEE.h"
#cgo LDFLAGS: -I../go-with-intel-sgx/App -L. -ltee 
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
    C.testMain()
    // test()
}