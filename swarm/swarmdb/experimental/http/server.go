package main

import (
	"github.com/ethereum/go-ethereum/crypto"

	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	swarmdb "github.com/ethereum/go-ethereum/swarmdb"
	"github.com/rs/cors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type HTTPServer struct {
	swarmdb    *swarmdb.SwarmDB
	listener   net.Listener
	keymanager swarmdb.KeyManager
	//lock       sync.Mutex
}

type SwarmDBReq struct {
	protocol string
	owner string
	table    string
	key      string
}

type HttpErrorResp struct {
	ErrorCode string `json:"errorcode,omitempty"`
	ErrorMsg  string `json:"errormsg,omitepty"`
}

func parsePath(path string) (swdbReq SwarmDBReq, err error) {
	pathParts := strings.Split(path, "/")
	if len(pathParts) < 2 {
		return swdbReq, fmt.Errorf("Invalid Path")
	} else {
		for k,v := range pathParts {
			switch k {
				case 1:
				swdbReq.protocol = v 
					
				case 2:
				swdbReq.owner = v

				case 3:
				swdbReq.table = v

				case 4:
				swdbReq.key = v
			}
		}
	}
	return swdbReq, nil
}

func StartHttpServer(config *swarmdb.SWARMDBConfig) {
	fmt.Println("\nstarting http server")
	httpSvr := new(HTTPServer)
	httpSvr.swarmdb = swarmdb.NewSwarmDB()
	km, errkm := swarmdb.NewKeyManager(config)
	if errkm != nil {
		//return errkm
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

	fmt.Printf("\nRunning ListenAndServe")
	fmt.Printf("\nListening on %s and port %d\n", config.ListenAddrHTTP, config.PortHTTP)
	addr := net.JoinHostPort(config.ListenAddrHTTP, strconv.Itoa(config.PortHTTP))
	//go http.ListenAndServe(config.Addr, hdlr)
	log.Fatal(http.ListenAndServe(addr, hdlr))
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if r.Method == "OPTIONS" {
		return
	}

	encAuthString := r.Header["Authorization"]
	var vUser *swarmdb.SWARMDBUser
	var errVerified error
	if len(encAuthString) == 0 {
		us := []byte("Hello, world!")
		msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(us), us)
		msg_hash := crypto.Keccak256([]byte(msg))
		fmt.Printf("\nMessage Hash: [%s][%x]", msg_hash, msg_hash)

		pa, _ := s.keymanager.SignMessage(msg_hash)
		fmt.Printf("\nUser: [%s], SignedMsg: [%x]", us, pa)
		vUser, errVerified = s.keymanager.VerifyMessage(msg_hash, pa)
	} else {
		encAuthStringParts := strings.SplitN(encAuthString[0], " ", 2)

		decAuthString, err := base64.StdEncoding.DecodeString(encAuthStringParts[1])
		if err != nil {
			return
		}

		decAuthStringParts := strings.SplitN(string(decAuthString), ":", 2)
		ethAddr := decAuthStringParts[0]
		ethAddrSigned := decAuthStringParts[1]

		ethAddrBytes, errEthAddr := hex.DecodeString(ethAddr)
		if errEthAddr != nil {
			fmt.Printf("ERR decoding eth Address:[%s]\n", ethAddrBytes)
		}

		ethAddrSignedBytes, errEthAddrSigned := hex.DecodeString(ethAddrSigned)
		if errEthAddrSigned != nil {
			fmt.Printf("ERR decoding response:[%s]\n", ethAddrSignedBytes)
		}
		vUser, errVerified = s.keymanager.VerifyMessage(ethAddrBytes, ethAddrSignedBytes)
		if errVerified != nil {
			fmt.Printf("\nError: %s", errVerified)
		}
	}
	verifiedUser := vUser

	//fmt.Println("HTTP %s request URL: '%s', Host: '%s', Path: '%s', Referer: '%s', Accept: '%s'", r.Method, r.RequestURI, r.URL.Host, r.URL.Path, r.Referer(), r.Header.Get("Accept"))
	swReq, _ := parsePath(r.URL.Path)

	var dataReq swarmdb.RequestOption
	var reqJson []byte
	if swReq.protocol != "swarmdb:" {
		//Invalid Protocol: Throw Error
		//fmt.Fprintf(w, "The protocol sent in: %s is invalid | %+v\n", swReq.protocol, swReq)
	} else {
		var err error
		if r.Method == "GET" {
			fmt.Fprintf(w, "Processing [%s] protocol request with Body of () \n", swReq.protocol)
			dataReq.RequestType = "Get"
			dataReq.Table = swReq.table
			dataReq.Key = swReq.key
			reqJson, err = json.Marshal(dataReq)
			if err != nil {
			}
		} else if r.Method == "POST" {
			bodyContent, _ := ioutil.ReadAll(r.Body)
			reqJson = bodyContent
			fmt.Printf("\nBODY Json: %s", reqJson)

			var bodyMapInt interface{}
			json.Unmarshal(bodyContent, &bodyMapInt)
			//fmt.Println("\nProcessing [%s] protocol request with Body of (%s) \n", swReq.protocol, bodyMapInt)
			//fmt.Fprintf(w, "\nProcessing [%s] protocol request with Body of (%s) \n", swReq.protocol, bodyMapInt)
			bodyMap := bodyMapInt.(map[string]interface{})
			if reqType, ok := bodyMap["requesttype"]; ok {
				dataReq.RequestType = reqType.(string)
				if dataReq.RequestType == "CreateTable" {
					dataReq.TableOwner = verifiedUser.Address //bodyMap["tableowner"].(string);
				} else if dataReq.RequestType == "Query" {
					dataReq.TableOwner = swReq.table 
					//Don't pass table for now (rely on Query parsing)
					if rq, ok := bodyMap["rawquery"]; ok {
						dataReq.RawQuery = rq.(string)
						reqJson, err = json.Marshal(dataReq)
						if err != nil {
						}
					} else {
						//Invalid Query Request: rawquery missing
					}
				} else if dataReq.RequestType == "Put" {
					dataReq.Table = swReq.table
					dataReq.TableOwner = swReq.owner 
					if row, ok := bodyMap["row"]; ok {
						//rowObj := make(map[string]interface{})
						//_ = json.Unmarshal([]byte(string(row.(map[string]interface{}))), &rowObj)
						newRow := swarmdb.Row{Cells: row.(map[string]interface{})}
						dataReq.Rows = append(dataReq.Rows, newRow)
					}
					reqJson, err = json.Marshal(dataReq)
					if err != nil {
					}
				}
			} else {
				fmt.Fprintf(w, "\nPOST operations require a requestType, (%+v), (%s)", bodyMap, bodyMap["requesttype"])
			}
		}
		//Redirect to SelectHandler after "building" GET RequestOption
		//fmt.Printf("Sending this JSON to SelectHandler (%s) and Owner=[%s]", reqJson, keymanager.WOLKSWARMDB_ADDRESS)
		response, errResp := s.swarmdb.SelectHandler(verifiedUser, string(reqJson))
		if errResp != nil {
			fmt.Printf("\nResponse resulted in Error: %s", errResp)
			httpErr := &HttpErrorResp{ErrorCode: "TBD", ErrorMsg: errResp.Error()}
			jHttpErr, _ := json.Marshal(httpErr)
			fmt.Fprint(w, string(jHttpErr))
		} else {
			fmt.Fprintf(w, response)
		}
	}
}

func main() {
	fmt.Println("Launching server...")

	// start swarm http proxy server
	config, _ := swarmdb.LoadSWARMDBConfig(swarmdb.SWARMDBCONF_FILE)
	StartHttpServer(&config)
	fmt.Println("\nAfter StartHttpServer Addr")
}
