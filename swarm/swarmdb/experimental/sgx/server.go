package main

import (
	"fmt"
	"net"
	"time"
	"math/rand"
	"os"
	"encoding/hex"
	"strings"
	swarmdb "github.com/ethereum/go-ethereum/swarmdb"
	"io"
	"bufio"
	common "github.com/ethereum/go-ethereum/swarmdb"
	"github.com/ethereum/go-ethereum/swarmdb/keymanager"
	"sync"
)

type ServerConfig struct {
	Addr string
	Port string
}

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
	keymanager keymanager.KeyManager
	lock       sync.Mutex
}

const (
	CONN_HOST = "127.0.0.1" // telnet 10.128.0.7 8501
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

func (client *Client) read() {
	for {
		_, err := client.reader.ReadString('\n')
		if err == io.EOF {
			client.conn.Close()
			break
		}
		if err != nil {
			//
		}
		// not implemented yet
		//resp := svr.swarmdb.SelectHandler(data)
		//fmt.Fprintf(client.writer, resp)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn, svr *TCPIPServer) {
	// generate a random 32 byte challenge (64 hex chars)
	// challenge = "27bd4896d883198198dc2a6213957bc64352ea35a4398e2f47bb67bffa5a1669"
	challenge := RandStringRunes(64)

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

	fmt.Printf("accepted connection [%s]\n", challenge);
	// Make a buffer to hold incoming data.
	//buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	// reqLen, err := conn.Read(buf)
	resp, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	} else {
		resp = strings.Trim(resp, "\r")
		resp = strings.Trim(resp, "\n")
	}

	// this should be the signed challenge, verify using valid_response
	challenge_bytes, err2 := hex.DecodeString(challenge)
	if err2 != nil {
		fmt.Printf("ERR decoding challenge:[%s]\n", challenge)
	}
	// resp = "6b1c7b37285181ef74fb1946968c675c09f7967a3e69888ee37c42df14a043ac2413d19f96760143ee8e8d58e6b0bda4911f642912d2b81e1f2834814fcfdad700"
	// fmt.Printf("BUF %d: %v\n", len([]byte(resp)), []byte(resp))

	response_bytes, err3 := hex.DecodeString(resp)
	// fmt.Printf("Response: [%d] %s \n", len(response_bytes), resp);
	if err3 != nil {
		fmt.Printf("ERR decoding response:[%s]\n", resp)
	}
	
	verified, err := svr.keymanager.VerifyMessage(challenge_bytes, response_bytes)
	if err != nil {
		resp = "ERR"
	}  else if verified {
		resp = "VALID"
	} else {
		resp = "INVALID"
	}
	fmt.Printf("%s C: %x R: %x\n", resp, challenge_bytes, response_bytes);
	fmt.Fprintf(writer, resp)
	writer.Flush()
	if ( resp == "VALID0" ) {
		go client.read()
	} else {
	}
	conn.Close()
	// Close the connection when you're done with it.
}

func StartTCPIPServer(swarmdb *common.SwarmDB, config *ServerConfig) (err error) {
	sv := new(TCPIPServer)
	sv.swarmdb = swarmdb
	km, errkm := keymanager.NewKeyManager(keymanager.PATH, keymanager.WOLKSWARMDB_ADDRESS, keymanager.WOLKSWARMDB_PASSWORD)
	if errkm != nil {
		return err
	} else {
		sv.keymanager = km
	}

	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	// l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", config.Port))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	} else {
		fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
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
		go handleRequest(conn,  sv)
	}
}


func main() {
	fmt.Println("Launching server...")
	swdb := swarmdb.NewSwarmDB()
	tcpaddr := net.JoinHostPort("127.0.0.1", "2000")
	StartTCPIPServer(swdb, &ServerConfig{
		Addr: tcpaddr,
		Port: "2000",
	})
}




