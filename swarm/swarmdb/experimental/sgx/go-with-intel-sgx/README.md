# Go with Intel SGX 
[![license](https://img.shields.io/github/license/mashape/apistatus.svg?style=flat-square)]()

There are many examples of cgo which connects C/C++ source codes to native Go program. And I also saw many sample codes written for Intel SGX. But I couldn't find a combination of both cgo and Intel SGX. So I made it.

This example would be useful to those who want to utilize Intel SGX using Go Lang.

If you are curious about how it works, check [cgo target in Makefile](https://github.com/rupc/go-with-intel-sgx/blob/master/Makefile#L214)

# How to test
```
source $SGX_SDK/environment # not needed when you already have it
git clone https://github.com/rupc/go-with-intel-sgx
cd go-with-intel-sgx
make cgo
```

# Features
Following features are demonstrated, running inside enclave.
- Monotonic counter
- ECDSA: private/public key generation, signing, verifying
- SHA256

# Thanks to
I expand [hello-enclave](https://github.com/digawp/hello-enclave) to build cgo-sgx connection
