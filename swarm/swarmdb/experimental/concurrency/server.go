package server

import (
	"bufio"
	//"bytes"
	//"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"github.com/ethereum/go-ethereum/log"
	"net"
	"os"
	common "github.com/ethereum/go-ethereum/swarmdb"
	"github.com/ethereum/go-ethereum/swarmdb/keymanager"
	"sync"
)

type IncomingInfo struct {
	Data    string
	Address string
}

type ServerConfig struct {
	Addr string
	Port string
}

type Client struct {
	conn     net.Conn
	incoming chan *common.IncomingInfo
	outgoing chan string
	reader   *bufio.Reader
	writer   *bufio.Writer
	table    *common.Table // holds ownerID, tableName
}

type TCPIPServer struct {
	swarmdb  *common.SwarmDB
	listener net.Listener
	keymanager keymanager.KeyManager
	conn     chan net.Conn
	incoming chan *common.IncomingInfo
	outgoing chan string
	clients  []*Client
	lock     sync.Mutex
}

func NewTCPIPServer(swarmdb *common.SwarmDB, l net.Listener) *TCPIPServer {
	sv := new(TCPIPServer)
	km, errkm := keymanager.NewKeyManager(keymanager.PATH, keymanager.WOLKSWARMDB_ADDRESS, keymanager.WOLKSWARMDB_PASSWORD)
	if errkm != nil {
	} else {
		sv.keymanager = km
	}
	sv.listener = l
	sv.clients = make([]*Client, 0)
	sv.conn = make(chan net.Conn)

	sv.incoming = make(chan *common.IncomingInfo)
	sv.outgoing = make(chan string)
	sv.swarmdb = swarmdb
	return sv
}

func StartTCPIPServer(swarmdb *common.SwarmDB, config *ServerConfig) {
	log.Debug(fmt.Sprintf("tcp StartTCPIPServer"))

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", config.Port))
	log.Debug(fmt.Sprintf("tcp StartTCPIPServer with %s", config.Port))

	svr := NewTCPIPServer(swarmdb, l)
	if err != nil {
		//log.Fatal(err)
		log.Debug(fmt.Sprintf("err"))
	}
	//defer svr.listener.Close()

	svr.listen()
	for {
		conn, err := svr.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		challenge := RandStringRunes(64)
		s := fmt.Sprintf(`%s\n`, challenge)
		conn.Write([]byte(s))
		svr.conn <- conn
	}
	if err != nil {
		//	log.Fatal(err)
		log.Debug(fmt.Sprintf("err"))
	}
	defer svr.listener.Close()
}

func newClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)
	client := &Client{
		conn:     connection,
		incoming: make(chan *common.IncomingInfo),
		outgoing: make(chan string),
		reader:   reader,
		writer:   writer,
		//databases: make(map[string]map[string]*common.Database),
	}
	go client.read()
	//go client.write()
	return client
}

func (client *Client) read() {
	for {
		line, err := client.reader.ReadString('\n')
		if err == io.EOF {
			client.conn.Close()
			break
		}
		if err != nil {
			////////
		}
		incoming := new(common.IncomingInfo)
		incoming.Data = line
		incoming.Address = client.conn.RemoteAddr().String()
		//client.incoming <- line
		client.incoming <- incoming
		fmt.Printf("[%s]Read:%s", client.conn.RemoteAddr(), line)
	}
}
func (client *Client) write() {
	for data := range client.outgoing {
		client.writer.WriteString(data)
		//client.writer.Write(data)
		client.writer.Flush()
		fmt.Printf("[%s]Write:%s\n", client.conn.RemoteAddr(), data)
	}
}

func (svr *TCPIPServer) addClient(conn net.Conn) {
	fmt.Printf("\nConnection Added")
	fmt.Fprintf(conn, "Your Connection Added\n")
	client := newClient(conn)
	/// this one is not good. need to change it
	svr.clients = append(svr.clients, client)
	go func() {
		for {
			svr.incoming <- <-client.incoming
			client.outgoing <- <-svr.outgoing
		}
	}()
}

func RandStringRunes(n int) string {
	// var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var letterRunes = []rune("0123456789abcdef")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (svr *TCPIPServer) TestAddClient(owner string, tablename string, primary string) {
	//testConn := svr.NewConnection()
	client := newClient(nil) //testConn)
	client.table = svr.swarmdb.NewTable(owner, tablename)
	//client.table.SetPrimary( primary )
	svr.clients = append(svr.clients, client)
}

func (svr *TCPIPServer) listen() {
	go func() {
		for {
			select {
			case conn := <-svr.conn:
				svr.addClient(conn)
			case data := <-svr.incoming:
				fmt.Printf("\nIncoming Data [%+v]", data)
				
				verified, err := svr.keymanager.VerifyMessage([]byte(data.Data), []byte(data.Data))
				if err != nil || !verified {
				
				} else {
					fmt.Fprintf(svr.clients[0].conn, "ok")
				}

				resp := svr.swarmdb.SelectHandler(data)
				fmt.Fprintf(svr.clients[0].conn, resp)
				svr.outgoing <- resp
			}
		}
	}()
}

func (svr *TCPIPServer) NewConnection() (err error) {
	ownerID := "owner1"
	tableName := "testtable"
	svr.swarmdb.NewTable(ownerID, tableName)

	// svr.table = table

	return nil
}
