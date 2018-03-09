package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	common "github.com/ethereum/go-ethereum/swarmdb"
	swarmdb "github.com/ethereum/go-ethereum/swarmdb"
	"io"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	svr    *TCPIPServer
	table  *common.Table // holds ownerID, tableName
}

type TCPIPServer struct {
	swarmdb    *common.SwarmDB
	listener   net.Listener
	keymanager swarmdb.KeyManager
	lock       sync.Mutex
}

const (
	CONN_HOST = "127.0.0.1"
	CONN_PORT = 2000
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

// Handles incoming requests.
func handleRequest(conn net.Conn, svr *TCPIPServer) {
	// generate a random 50 char challenge (64 hex chars)
	challenge := RandStringRunes(50)
	// challenge := "Hello, world!"
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	client := &Client{
		conn:   conn,
		reader: reader,
		writer: writer,
		svr:    svr,
	}

	fmt.Fprintf(writer, "%s\n", challenge)
	writer.Flush()

	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(challenge), challenge)
	challenge_bytes := crypto.Keccak256([]byte(msg))

	resp, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	} else {
		resp = strings.Trim(resp, "\r")
		resp = strings.Trim(resp, "\n")
	}

	// this should be the signed challenge, verify using valid_response
	response_bytes, err3 := hex.DecodeString(resp)
	if err3 != nil {
		fmt.Printf("ERR decoding response:[%s]\n", resp)
	}
	u, err := svr.keymanager.VerifyMessage(challenge_bytes, response_bytes)
	if err != nil {
		conn.Close()
	} else {
		fmt.Printf("%s Server Challenge [%s]-ethsign->[%x] Client %d byte Response:[%s] \n", resp, challenge, challenge_bytes, len(response_bytes), resp)
		// fmt.Fprintf(writer, "OK\n")
		writer.Flush()
		for {
			str, err := client.reader.ReadString('\n')
			if err == io.EOF {
				// Close the connection when done
				conn.Close()
				break
			}
			if true {
				resp, err := svr.swarmdb.SelectHandler(u, str)
				if err != nil {
					s := fmt.Sprintf("ERR: %s\n", err)
					fmt.Printf(s)
					writer.WriteString(s)
					writer.Flush()
				} else {
					fmt.Printf("Read: [%s] Wrote: [%s]\n", str, resp)
					writer.WriteString(resp + "\n")
					writer.Flush()
					// 					fmt.Fprintf(client.writer, resp + "\n")
				}
			} else {
				writer.WriteString("OK\n")
				writer.Flush()
			}
		}
	}
}

func StartTCPIPServer(sdb *common.SwarmDB, conf *swarmdb.SWARMDBConfig) (err error) {

	sv := new(TCPIPServer)
	sv.swarmdb = sdb
	km, errkm := swarmdb.NewKeyManager(conf)
	if errkm != nil {
		return err
	} else {
		sv.keymanager = km
	}

	// Listen for incoming connections.
	host := CONN_HOST
	port := CONN_PORT
	if len(conf.ListenAddrTCP) > 0 {
		host = conf.ListenAddrTCP
	}
	if conf.PortTCP > 0 {
		port = conf.PortTCP
	}

	host_port := fmt.Sprintf("%s:%d", host, port)
	l, err := net.Listen(CONN_TYPE, host_port)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	} else {
		fmt.Println("Listening on " + host_port)
	}
	// Close the listener when the application closes.
	defer l.Close()

	// sv.listener = l

	// generate "truly" random strings
	rand.Seed(time.Now().UnixNano())
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn, sv)
	}
}

func main() {
	fmt.Println("Launching server...")
	swdb := swarmdb.NewSwarmDB()
	config, _ := swarmdb.LoadSWARMDBConfig(swarmdb.SWARMDBCONF_FILE)
	StartTCPIPServer(swdb, &config)
}
