
package main

/*
#include "TEE.h"
#cgo LDFLAGS: -IApp -L. -ltee 
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