package main

/*
#include "./App/TEE.h"
#cgo LDFLAGS: -I./App -L. -ltee 
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

//export test
func test() {
    fmt.Printf("intel sgx hello\n")
}

func main() {
	
    //test()
    
/*  
    //***************************************************** SHA256 *****************************************************  
    // http://geekwentfreak.com/posts/golang/cgo_pass_receive_string_c/ 
    cstr := C.CString("TestThisSGX")
    defer C.free(unsafe.Pointer(cstr))
    cString := C.getSha256(cstr)          // hash: 18497686A320B7DA753F7E18C58C4F2E18089D5816FDE68858878D22C8237E36
    gostr := C.GoString(cString)
    fmt.Println("Received hash (string) from C: " + gostr)
    //***************************************************** SHA256 ****************************************************   
*/
    //***************************************************** sgx_ecc256_create_key_pair ****************************************************
    // swarm.wolk.com/sgx/go-with-intel-sgx/Enclave/Enclave.cpp    
    privateKey := C.CString("0000000000000000000000000000000000000000000000000000000000000000")
    defer C.free(unsafe.Pointer(privateKey))
    publicKeyGX := C.CString("0000000000000000000000000000000000000000000000000000000000000000")
    defer C.free(unsafe.Pointer(publicKeyGX))    
    publicKeyGY := C.CString("0000000000000000000000000000000000000000000000000000000000000000")
     defer C.free(unsafe.Pointer(publicKeyGY))
        
    C.ecc256CreateKeyPair(privateKey, publicKeyGX, publicKeyGY)
    
    gostrPrivateKey := C.GoString(privateKey)
    fmt.Println("Received gostrPrivateKey from C: " + gostrPrivateKey)
    gostrPublicKeyGX := C.GoString(publicKeyGX)
    fmt.Println("Received gostrPublicKeyGX from C: " + gostrPublicKeyGX)    
    gostrPublicKeyGY := C.GoString(publicKeyGY)
    fmt.Println("Received gostrPublicKeyGY from C: " + gostrPublicKeyGY)    
    //***************************************************** sgx_ecc256_create_key_pair ****************************************************    
    
   //***************************************************** sgx_ecdsa_sign ****************************************************     
    
    
   // C.ecdsaSign(privateKey)
    
    
    
   //***************************************************** sgx_ecdsa_sign ****************************************************     
    
    
    test()
}
