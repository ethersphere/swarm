# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

GOBIN = $(shell pwd)/build/bin

swarm:
	build/env.sh go run -mod=vendor build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

alltools:
	build/env.sh go run -mod=vendor build/ci.go install ./cmd/...

# Wrap go modules vendor command to copy forked cgo libraries
# from go module cache and correct their file permissons.
.PHONY: vendor
vendor: export GO111MODULE=on
vendor:
	@go mod vendor
	@cp -rf "$(shell GO111MODULE=on go list -f {{.Dir}} github.com/karalabe/hid)/hidapi" vendor/github.com/karalabe/hid/hidapi
	@chmod -R u+w vendor/github.com/karalabe/hid/hidapi
	@cp -rf "$(shell GO111MODULE=on go list -f {{.Dir}} github.com/karalabe/hid)/libusb" vendor/github.com/karalabe/hid/libusb
	@chmod -R u+w vendor/github.com/karalabe/hid/libusb
	@cp -rf "$(shell GO111MODULE=on go list -f {{.Dir}} github.com/ethereum/go-ethereum/crypto/secp256k1)/libsecp256k1" vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1
	@chmod -R u+w vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1
