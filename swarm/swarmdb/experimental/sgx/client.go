package main

import (
	"bufio"
	"fmt"
	"github.com/ethereum/go-ethereum/swarmdb/keymanager"
	"math/rand"
	"net"
	"encoding/hex"
	"os"
	"strings"
)

const (
	CONN_HOST = "127.0.0.1"
	CONN_PORT = "2000"
	CONN_TYPE = "tcp"
)

func RandStringRunes(n int) string {
	var letterRunes = []rune("0123456789abcdef")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}



func main() {
	// open a TCP connection to ip port
	km, err := keymanager.NewKeyManager(keymanager.PATH, keymanager.WOLKSWARMDB_ADDRESS, keymanager.WOLKSWARMDB_PASSWORD)

	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Printf("Connection Error: %v\n", err)
		os.Exit(0)
	}
	fmt.Printf("Opened connection: reading string...")
	challenge, _ := bufio.NewReader(conn).ReadString('\n')
	challenge = strings.Trim(challenge, "\n")
	// challenge = "27bd4896d883198198dc2a6213957bc64352ea35a4398e2f47bb67bffa5a1669"
	challenge_bytes, _ := hex.DecodeString(challenge)
	sig, err := km.SignMessage(challenge_bytes)
	if err != nil {
		fmt.Printf("Err %s\n", err)
	} else {
		fmt.Printf("Challenge: [%x] Sig:[%x]\n", challenge_bytes, sig)
	}
// response = "6b1c7b37285181ef74fb1946968c675c09f7967a3e69888ee37c42df14a043ac2413d19f96760143ee8e8d58e6b0bda4911f642912d2b81e1f2834814fcfdad700"
//	response := generate_challenge_response(challenge)
	response := fmt.Sprintf("%x", sig)
	fmt.Printf("challenge:[%v] response:[%v]\n", challenge, response)
	fmt.Fprintf(conn, response+"\n")
	message2, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Printf("%s\n", message2)
}
