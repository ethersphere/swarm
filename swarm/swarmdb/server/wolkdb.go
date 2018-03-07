package main

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	swarmdb "github.com/ethereum/go-ethereum/swarmdb"
	sdbc "github.com/ethereum/go-ethereum/swarmdb/swarmdbcommon"
	"github.com/rs/cors"
	"io"
	"io/ioutil"
	logger "log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	svr    *TCPIPServer
	table  *swarmdb.Table // holds ownerID, tableName
}

type TCPIPServer struct {
	swarmdb    *swarmdb.SwarmDB
	listener   net.Listener
	keymanager swarmdb.KeyManager
	lock       sync.Mutex
}

func RandStringRunes(n int) string {
	var letterRunes = []rune("0123456789abcdef")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type HTTPServer struct {
	swarmdb    *swarmdb.SwarmDB
	listener   net.Listener
	keymanager swarmdb.KeyManager
	//lock       sync.Mutex
}

type SwarmDBReq struct {
	protocol string
	owner    string
	database string
	table    string
	key      string
}

type ChallengeVersionPair struct {
	Challenge     string `json:"challenge,omitempty"`
	ServerVersion string `json:"serverversion,omitempty"`
}

type ResponseVersionPair struct {
	Response      string `json:"response,omitempty"`
	ClientVersion string `json:"clientversion,omitempty"`
	ClientName    string `json:"clientname,omitempty"`
}

func (rvp *ChallengeVersionPair) validClientVersion(cvp ResponseVersionPair) (ok bool) {
	return true
}

func buildErrorResp(err error) string {
	var respObj sdbc.SWARMDBResponse
	wolkErr, ok := err.(*sdbc.SWARMDBError)
	if !ok {
		return (`{ "errorcode":-1, "errormessage":"UNKNOWN ERROR"}`) //TODO: Make Default Error Handling
	}
	if wolkErr.ErrorCode == 0 { //FYI: default empty int is 0. maybe should be a pointer.  //TODO this is a hack with what errors are being returned right now
		//fmt.Printf("wolkErr.ErrorCode doesn't exist\n")
		respObj.ErrorCode = 888
		respObj.ErrorMessage = err.Error()
	} else {
		respObj.ErrorCode = wolkErr.ErrorCode
		respObj.ErrorMessage = wolkErr.ErrorMessage
	}
	jbyte, jErr := json.Marshal(respObj)
	if jErr != nil {
		//fmt.Printf("Error: [%s] [%+v]", jErr.Error(), respObj)
		return `{ "errorcode":-1, "errormessage":"UNKNOWN ERROR"}` //TODO: Make Default Error Handling
	}
	log.Debug(fmt.Sprintf("Internal Error Message: %s", wolkErr.Error()))
	jstr := string(jbyte)
	return jstr
}

// Handles incoming TCPIP requests.
func handleTcpipRequest(conn net.Conn, svr *TCPIPServer) {
	// generate a random 50 char challenge (64 hex chars)
	var cvp ChallengeVersionPair
	cvp.Challenge = RandStringRunes(50)
	cvp.ServerVersion = swarmdb.SWARMDBVersion

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	client := &Client{
		conn:   conn,
		reader: reader,
		writer: writer,
		svr:    svr,
	}
	challenge, errCh := json.Marshal(cvp)
	if errCh != nil {
		log.Debug(fmt.Sprintf("ERROR marshalling? %s", errCh.Error()))
		//Todo: error
	}
	log.Debug(fmt.Sprintf("SENDING BACK %s from %+v", challenge, cvp))
	fmt.Fprintf(writer, "%s\n", challenge)
	writer.Flush()

	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(cvp.Challenge), cvp.Challenge)
	challenge_bytes := crypto.Keccak256([]byte(msg))

	var swErr sdbc.SWARMDBError
	rvp, err := reader.ReadString('\n')
	if err != nil {
		swErr.SetError(fmt.Sprintf("Problem reading RAW TCPIP input (%s).  ERROR:[%s]", rvp, err.Error()))
		swErr.ErrorCode = 478
		swErr.ErrorMessage = "Unable to Parse Response to Challenge"
		log.Error(swErr.Error())
		tcpJson := buildErrorResp(&swErr)
		writer.WriteString(tcpJson)
		writer.Flush()
		return
	}

	var resp ResponseVersionPair
	errRvpUnmarshal := json.Unmarshal([]byte(rvp), &resp)
	if errRvpUnmarshal != nil {
		swErr.SetError(fmt.Sprintf("Unable to parse Response Value Pair sent to server: %s", errRvpUnmarshal.Error()))
		swErr.ErrorCode = 478
		swErr.ErrorMessage = "Unable to Parse RAW TCP Input"
		log.Error(swErr.Error())
		tcpJson := buildErrorResp(&swErr)
		writer.WriteString(tcpJson)
		writer.Flush()
		return
	}
	resp.Response = strings.Trim(resp.Response, "\r")
	resp.Response = strings.Trim(resp.Response, "\n")
	log.Debug(fmt.Sprintf("handleTcpipRequest response %v\n", resp))

	// this should be the signed challenge, verify using valid_response
	response_bytes, errDecoding := hex.DecodeString(resp.Response)
	if errDecoding != nil {
		swErr.SetError(fmt.Sprintf("Problem decoding TCPIP input.  ERROR:[%s]", errDecoding.Error()))
		log.Error(swErr.Error())
		swErr.ErrorCode = 422
		swErr.ErrorMessage = fmt.Sprintf("Unable to Decode Response sent [%s]", resp.Response)
		tcpJson := buildErrorResp(&swErr)
		writer.WriteString(tcpJson)
		writer.Flush()
		return
	}

	log.Debug(fmt.Sprintf("[wolkdb:handleTcpipRequest] Valid Response from [%s] [%s]", resp.ClientName, resp.ClientVersion))

	if !cvp.validClientVersion(resp) {
		swErr.SetError(fmt.Sprintf("Client Version [%s] of %s is incompatible with SWARMDB Server version [%s]", resp.ClientVersion, resp.ClientName, cvp.ServerVersion))
		swErr.ErrorCode = 478
		swErr.ErrorMessage = "Unable to Parse RAW TCP Input"
		log.Error(swErr.Error())
		tcpJson := buildErrorResp(&swErr)
		writer.WriteString(tcpJson)
		writer.Flush()
		return
	}

	u, err := svr.keymanager.VerifyMessage(challenge_bytes, response_bytes)
	if err != nil {
		log.Debug(fmt.Sprintf("ERROR: %s", err.Error()))
		swErr.SetError(fmt.Sprintf("Unable to verify challenge response. Please check your PRIVATE_KEY: [%s]", err.Error()))
		swErr.ErrorCode = 484
		swErr.ErrorMessage = "Challenge Response Failed.  Please check your PRIVATE_KEY"
		log.Error(swErr.Error())
		tcpJson := buildErrorResp(err)
		writer.WriteString(tcpJson)
		writer.Flush()
		conn.Close()
	} else {
		log.Debug("%s Server Challenge [%s]-ethsign->[%x] Client %d byte Response:[%s] \n", resp, challenge, challenge_bytes, len(response_bytes), resp.Response)
		writer.Flush()
		for {
			str, err := client.reader.ReadString('\n')
			if err == io.EOF {
				//TODO: return a TCPIP error response
				// Close the connection when done
				conn.Close()
				break
			}
			if true {
				log.Debug("[main:handleTcpipRequest] sending [%s]", str)
				if slResp, err := svr.swarmdb.SelectHandler(u, string(str)); err != nil {
					log.Debug(fmt.Sprintf("ERROR: %+v", err))
					tcpJson := buildErrorResp(err)
					log.Debug(fmt.Sprintf("Read: [%s] Wrote: [%s]\n", str, tcpJson))
					_, err := writer.WriteString(tcpJson + "\n")
					if err != nil {
						fmt.Printf("writer err: %v\n", err)
						//TODO handle if writestring has err
					}
					writer.Flush()
				} else {
					log.Debug(fmt.Sprintf("Read: [%s] Wrote: [%s]\n", str, slResp.Stringify()))
					_, err := writer.WriteString(slResp.Stringify() + "\n")
					if err != nil {
						fmt.Printf("writer err: %v\n", err)
						//TODO handle if writestring has err
					}
					writer.Flush()
				}
			} else {
				writer.WriteString("OK\n")
				writer.Flush()
			}
		}
	}
}

func StartTcpipServer(sdb *swarmdb.SwarmDB, conf *swarmdb.SWARMDBConfig) (err error) {
	sv := new(TCPIPServer)
	sv.swarmdb = sdb
	km, errkm := swarmdb.NewKeyManager(conf)
	if errkm != nil {
		log.Error(errkm.Error())
		return errkm
	} else {
		sv.keymanager = km
	}

	host := swarmdb.SWARMDBCONF_LISTENADDR
	port := swarmdb.SWARMDBCONF_PORTTCP

	// Listen for incoming connections.
	if len(conf.ListenAddrTCP) > 0 {
		host = conf.ListenAddrTCP
	}
	if conf.PortTCP > 0 {
		port = conf.PortTCP
	}

	//TODO: Do we want default host/port if not in config?

	host_port := fmt.Sprintf("%s:%d", host, port)
	l, err := net.Listen("tcp", host_port)

	var swErr sdbc.SWARMDBError
	if err != nil {
		swErr.SetError(fmt.Sprintf("Error trying to listen (tcp) on host/port [%s].  ERROR:[%s]", host_port, err))
		log.Error(swErr.Error())
		return err //TODO: investigate why returning swErr fails
		os.Exit(1) //TODO: should we exit?
	} else {
		log.Debug("TCPIP Server Listening on " + host_port)
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
		go handleTcpipRequest(conn, sv)
	}
}

func parsePath(path string) (swdbReq SwarmDBReq, err error) {
	var swErr sdbc.SWARMDBError
	pathParts := strings.Split(path, "/")
	if len(pathParts) < 2 {
		swErr = sdbc.SWARMDBError{ErrorCode: -1, ErrorMessage: "Request URL invalid"}
		swErr.SetError("Invalid Path in Request URL")
		return swdbReq, &swErr
	} else {
		for k, v := range pathParts {
			switch k {
			case 1:
				swdbReq.owner, swdbReq.database, err = parseOwnerDB(v)
				if err != nil {
					return swdbReq, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("Invalid Owner/ENS path passed in [%s]", v))
				}
			case 2:
				swdbReq.table = v
			case 3:
				swdbReq.key = v
			default:
				//TODO:
			}
		}
	}
	return swdbReq, nil
}

func parseOwnerDB(v string) (owner string, db string, err error) {
	vParts := strings.Split(v, ".")
	if len(vParts) < 3 {
		//return db, owner, &sdbc.SWARMDBError{ErrorCode: -1, ErrorMessage: "Owner portion of request invalid"}
		//TODO: robust error!
	}
	owner = fmt.Sprintf("%s.%s", vParts[len(vParts)-2], vParts[len(vParts)-1])
	var dbParts []string
	for k, v := range vParts {
		if k == len(vParts)-2 {
			break
		}
		dbParts = append(dbParts, v)
	}
	db = strings.Join(dbParts, ".")
	return owner, db, nil
}

func StartHttpServer(sdb *swarmdb.SwarmDB, config *swarmdb.SWARMDBConfig) {
	httpSvr := new(HTTPServer)
	httpSvr.swarmdb = sdb
	km, errkm := swarmdb.NewKeyManager(config)
	if errkm != nil {
		//return errkm
		retJson := buildErrorResp(errkm)
		fmt.Printf(retJson)
		//TODO: show error to client
	} else {
		httpSvr.keymanager = km
	}
	var allowedOrigins []string
	/*
	   for _, domain := range strings.Split(config.CorsString, ",") {
	*/
	allowedOrigins = append(allowedOrigins, "corsdomain")
	// }
	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"POST", "GET", "DELETE", "PATCH", "PUT"},
		MaxAge:         600,
		AllowedHeaders: []string{"*"},
	})
	//sk, pk := GetKeys()
	hdlr := c.Handler(httpSvr)

	log.Debug(fmt.Sprintf("HTTP Listening on %s and port %d", config.ListenAddrHTTP, config.PortHTTP))
	addr := net.JoinHostPort(config.ListenAddrHTTP, strconv.Itoa(config.PortHTTP))
	//go http.ListenAndServe(config.Addr, hdlr)
	logger.Fatal(http.ListenAndServe(addr, hdlr))
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if r.Method == "OPTIONS" {
		return
	}

	var swErr sdbc.SWARMDBError
	encAuthString := r.Header["Authorization"]
	var vUser *swarmdb.SWARMDBUser
	var errVerified error
	bodyContent, errReadBody := ioutil.ReadAll(r.Body)
	if errReadBody != nil {
		//TODO: Handle Reading Body error
		swErr.SetError(fmt.Sprintf("[wolkdb:ServeHTTP] Error Reading Request Body.[%s]", errReadBody.Error()))
		log.Error(swErr.Error())
		swErr.ErrorCode = 422
		swErr.ErrorMessage = fmt.Sprintf("Error Reading Request Body: [%s]", errReadBody.Error())
		retJson := buildErrorResp(&swErr)
		fmt.Fprint(w, retJson)
	}
	reqJson := bodyContent

	pathParts := strings.Split(r.URL.Path, "/")
	var logData []string
	switch pathParts[1] {
	case "swaplog":
		startts, err := strconv.Atoi(pathParts[2])
		if err != nil {
			//TODO: Error Handling
		}
		endts, err := strconv.Atoi(pathParts[3])
		if err != nil {
			//TODO: Error Handling
		}
		logData, err = s.swarmdb.GenerateSwapLog(int64(startts), int64(endts))
		if err != nil {
			//TODO: Error Handling
		}

		return
	case "buyerlog":
		startts, err := strconv.Atoi(pathParts[2])
		if err != nil {
			//TODO: Error Handling
		}
		endts, err := strconv.Atoi(pathParts[3])
		if err != nil {
			//TODO: Error Handling
		}
		logData, err = s.swarmdb.GenerateBuyerLog(int64(startts), int64(endts))
		if err != nil {
			//TODO: Error Handling
		}
		return
	case "farmerlog":
		startts, err := strconv.Atoi(pathParts[2])
		if err != nil {
			//TODO: Error Handling
		}
		endts, err := strconv.Atoi(pathParts[3])
		if err != nil {
			//TODO: Error Handling
		}
		logData, err = s.swarmdb.GenerateFarmerLog(int64(startts), int64(endts))
		if err != nil {
			//TODO: Error Handling
		}

		return
	case "ashrequest":
		chunkID, _ := hex.DecodeString(pathParts[2])
		seed, _ := hex.DecodeString(pathParts[3])
		auditIndex, err := strconv.Atoi(pathParts[4])
		if err != nil {
			//TODO: Error Handling
		}
		proofRequired := false
		if len(pathParts) > 4 {
			if pathParts[5] == "true" || pathParts[5] == "1" {
				proofRequired = true
			}
		}

		fmt.Printf("ChunkID:%x | seed:%x | ProofRequired:%t | Index: %d\n", chunkID, seed, proofRequired, int8(auditIndex))
		resp, err := s.swarmdb.GenerateAshResponse(chunkID, seed, proofRequired, int8(auditIndex))
		if err != nil {
			//TODO: Error Handling
		}
		output, _ := json.Marshal(resp)
		fmt.Printf("%s\n", string(output))
		return
	default:
	}
	log.Debug(fmt.Sprintf("Logdata", logData))
	swReq, err := parsePath(r.URL.Path)
	if err != nil {
		retJson := buildErrorResp(err)
		fmt.Fprint(w, retJson)
	}
	log.Debug(fmt.Sprintf("swReq [%+v]", swReq))

	if len(encAuthString) == 0 {
		//TODO: remove "backdoor"
		us := []byte(`{ "requesttype":"Put", "row":{"email":"rodney@wolk.com", "name":"Rodney F. Witcher", "age":370} }`)
		msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(us), us)
		msg_hash := crypto.Keccak256([]byte(msg))
		//fmt.Printf("\nMessage Hash: [%s][%x]", msg_hash, msg_hash)

		pa, _ := s.keymanager.SignMessage(msg_hash)
		//TODO: SignMessageError

		//fmt.Printf("\nUser: [%s], Msg Hash [%x], SignedMsg: [%x]\n", us, msg_hash, pa)
		vUser, errVerified = s.keymanager.VerifyMessage(msg_hash, pa)
		if errVerified != nil {
			//TODO: Show Error to Client
		}
	} else {
		bodyContentSeed := bodyContent
		if r.Method == "GET" {
			bodyContentSeed = []byte(fmt.Sprintf("%s%s%s", swReq.owner, swReq.table, swReq.key))
		}
		encAuthStringParts := strings.SplitN(encAuthString[0], " ", 2)
		decAuthString, err := base64.StdEncoding.DecodeString(encAuthStringParts[1])
		if err != nil {
			return
		}

		fmt.Printf("\nDecAuthString: [%x][%s]", decAuthString, decAuthString)
		decAuthStringParts := strings.SplitN(string(decAuthString), ":", 2)
		inputSignedMsg := decAuthStringParts[0]

		msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(bodyContentSeed), bodyContentSeed)
		msg_hash := crypto.Keccak256([]byte(msg))
		fmt.Printf("\nMessage Hash: [%s][%x]", msg_hash, msg_hash)

		decSignedMsg, errDecSignedMsg := hex.DecodeString(inputSignedMsg)
		if errDecSignedMsg != nil {
			swErr.SetError(fmt.Sprintf("[wolkdb:ServeHTTP] Error Decoding Signed Message.[%s] %s", inputSignedMsg, errDecSignedMsg.Error()))
			log.Error(swErr.Error())
			swErr.ErrorCode = 422
			swErr.ErrorMessage = fmt.Sprintf("Error Decoding Signed Message", errDecSignedMsg.Error())
			retJson := buildErrorResp(&swErr)
			fmt.Fprintf(w, retJson)
		}
		//fmt.Printf("\nSignedMsg: [%x][%s] | DecSignedMsg: [%x][%s]", signedMsg, signedMsg, decSignedMsg, decSignedMsg)

		vUser, errVerified = s.keymanager.VerifyMessage(msg_hash, decSignedMsg)
		if errVerified != nil {
			fmt.Printf("\nError: %s", errVerified)
			retJson := buildErrorResp(errVerified)
			fmt.Fprintf(w, retJson)
		}
	}
	verifiedUser := vUser
	var dataReq sdbc.RequestOption
	dataReq.Owner = swReq.owner
	dataReq.Database = swReq.database
	dataReq.Table = swReq.table
	dataReq.Key = swReq.key
	if r.Method == "GET" {
		//fmt.Fprintf(w, "Processing [%s] protocol request with Body of () \n", swReq.protocol)
		dataReq.RequestType = "Get"
		reqJson, err = json.Marshal(dataReq)
		if err != nil {
			//TODO: Return Error to Client
			swErr.SetError(fmt.Sprintf("[wolkdb:ServeHTTP] Error Marshaling request, %s", err.Error()))
			log.Error(swErr.Error())
			swErr.ErrorCode = 424
			swErr.ErrorMessage = fmt.Sprintf("Error Reading Request", err.Error())
			retJson := buildErrorResp(&swErr)
			fmt.Fprint(w, retJson)
		}
	} else if r.Method == "POST" {
		//fmt.Printf("\nBODY Json: %s", reqJson)

		var bodyMapInt interface{}
		json.Unmarshal(bodyContent, &bodyMapInt)
		log.Debug(fmt.Sprintf("Processing request with Body of (%s) \n", bodyMapInt))
		if bodyMap, ok := bodyMapInt.(map[string]interface{}); ok {
			if reqType, ok := bodyMap["requesttype"]; ok {
				dataReq.RequestType = reqType.(string)
				log.Debug(fmt.Sprintf("Table (%s) [%+v]", dataReq.Table, dataReq))
				if dataReq.Table == "" {
					log.Debug("Table not included in URL. Checking RequestBody")
					if tblBody, ok := bodyMap["table"]; ok {
						dataReq.Table = tblBody.(string)
					}
				}
				if dataReq.RequestType == "CreateTable" {
					bodyMap["owner"] = swReq.owner
					bodyMap["database"] = swReq.database
					//TODO: ValidateCreateTableRequest
					reqJson, err = json.Marshal(bodyMap)
					//TODO: error check
				} else if dataReq.RequestType == "Query" {
					//Don't pass table for now (rely on Query parsing)
					if rq, ok := bodyMap["query"]; ok {
						dataReq.RawQuery = rq.(string)
						reqJson, err = json.Marshal(dataReq)
						if err != nil {
							swErr.SetError(fmt.Sprintf("[wolkdb:ServeHTTP] Error Marshaling request, %s", err.Error()))
							log.Error(swErr.Error())
							swErr.ErrorCode = 424
							swErr.ErrorMessage = fmt.Sprintf("Error Reading Request", err.Error())
							retJson := buildErrorResp(&swErr)
							fmt.Fprint(w, retJson)
						}
					} else {
						//Invalid Query Request: query missing
						swErr.SetError(fmt.Sprintf("[wolkdb:ServeHTTP] Invalid Query Request.  Missing Query"))
						log.Error(swErr.Error())
						swErr.ErrorCode = 425
						swErr.ErrorMessage = fmt.Sprintf("Invalid Query Request. Missing query")
						retJson := buildErrorResp(&swErr)
						fmt.Fprint(w, retJson)
					}
				} else if dataReq.RequestType == "Put" {
					if row, ok := bodyMap["row"]; ok {
						newRow := sdbc.NewRow()
						newRow = row.(map[string]interface{})
						dataReq.Rows = append(dataReq.Rows, newRow)
					}
					reqJson, err = json.Marshal(dataReq)
					if err != nil {
						//TODO: Return Error to Client
						swErr.SetError(fmt.Sprintf("[wolkdb:ServeHTTP] Error Marshaling request, %s", err.Error()))
						log.Error(swErr.Error())
						swErr.ErrorCode = 424
						swErr.ErrorMessage = fmt.Sprintf("Error Reading Request", err.Error())
						retJson := buildErrorResp(&swErr)
						fmt.Fprintf(w, retJson)
					}
				} else {
					bodyMap["owner"] = swReq.owner
					bodyMap["database"] = swReq.database
					bodyMap["table"] = swReq.table
					reqJson, err = json.Marshal(bodyMap)
				}
			} else {
				swErr = sdbc.SWARMDBError{ErrorCode: 438, ErrorMessage: "Invalid Request Body -- Missing requesttype"}
				swErr.SetError(fmt.Sprintf("POST operations require a requestType, (%+v), (%s)", bodyMap, bodyMap["requesttype"]))
				retJson := buildErrorResp(&swErr)
				fmt.Fprint(w, retJson)
			}
		} else {
			swErr = sdbc.SWARMDBError{ErrorCode: 438, ErrorMessage: "Invalid Request Body"}
			swErr.SetError(fmt.Sprintf("Input Data Invalid [%v]", bodyMapInt))
			log.Debug(swErr.Error())
			retJson := buildErrorResp(&swErr)
			fmt.Fprint(w, retJson)
		}
	}
	if swErr.ErrorMessage == "" {
		//Redirect to SelectHandler after "building" GET RequestOption
		log.Debug(fmt.Sprintf("JSON sent in request [%s]", reqJson))
		response, errResp := s.swarmdb.SelectHandler(verifiedUser, string(reqJson))
		if errResp != nil {
			retJson := buildErrorResp(errResp)
			fmt.Fprint(w, retJson)
		} else {
			fmt.Fprintf(w, response.Stringify())
		}
	}
}

func main() {
	configFileLocation := flag.String("config", swarmdb.SWARMDBCONF_FILE, "Full path location to SWARMDB configuration file.")
	//TODO: store this somewhere accessible to be used later
	logLevelFlag := flag.Int("loglevel", 3, "Log Level Verbosity 1-6 (4 for debug)")
	version := flag.Bool("v", false, "Prints current SWARMDB version")
	flag.Parse()

	if *version {
		log.Debug(fmt.Sprintf("Working on version %s of SWARMDB Server\n", swarmdb.SWARMDBVersion))
		fmt.Printf("Working on version %s of SWARMDB Sever\n", swarmdb.SWARMDBVersion)
		os.Exit(0)
	}
	if _, err := os.Stat(*configFileLocation); os.IsNotExist(err) {
		log.Debug("Default config file missing.  Building ..")
		_, err := swarmdb.NewKeyManagerWithoutConfig(*configFileLocation, swarmdb.SWARMDBCONF_DEFAULT_PASSPHRASE)
		if err != nil {
			//TODO
		}
	}

	config, err := swarmdb.LoadSWARMDBConfig(*configFileLocation)
	if err != nil {
		log.Debug("The config file location provided [%s] is invalid.  Exiting ...", *configFileLocation)
		os.Exit(1)
	}

	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*logLevelFlag), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
	log.Debug(fmt.Sprintf("Starting SWARMDB (Version: %s) using [%s] and loglevel [%d]", swarmdb.SWARMDBVersion, *configFileLocation, *logLevelFlag))

	swdb, err := swarmdb.NewSwarmDB(config)
	if err != nil {
		panic(fmt.Sprintf("Cannot start: %s", err.Error()))
	}
	log.Debug("Trying to start HttpServer")
	go StartHttpServer(swdb, config)

	log.Debug("Trying to start TCPIP server...\n")
	StartTcpipServer(swdb, config)
}
