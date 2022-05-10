/**
* (C) 2019 Geotab Inc
* (C) 2019 Volvo Cars
*
* All files and artifacts in the repository at https://github.com/w3c/automotive-viss2
* are licensed under the provisions of the license provided by the LICENSE file in this repository.
*
**/

package utils

import (
	"encoding/json"
	"io/ioutil"

	"crypto/tls"
	"crypto/x509"
	"net/http"

	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const backendTermination = "internal-backend-termination"

func getWsClientIndex() int {
    freeIndex := -1
    for i := range WsClientIndexList {
        if (WsClientIndexList[i] == true) {
            WsClientIndexList[i] = false
            freeIndex = i
            break
        }
    }
    return freeIndex
}

func ReturnWsClientIndex(index int) {
    WsClientIndexList[index] = true
}

func ReadTransportSecConfig() {
	data, err := ioutil.ReadFile(trSecConfigPath + "transportSec.json")
	if err != nil {
		Info.Printf("ReadTransportSecConfig():%stransportSec.json error=%s", trSecConfigPath, err)
		secConfig.TransportSec = "no"
		return
	}
	err = json.Unmarshal(data, &secConfig)
	if err != nil {
		Error.Printf("ReadTransportSecConfig():Error unmarshal transportSec.json=%s", err)
		secConfig.TransportSec = "no"
		return
	}
	Info.Printf("ReadTransportSecConfig():secConfig.TransportSec=%s", secConfig.TransportSec)
}

func createRouterIdProperty(mgrId int, clientId int) string {
	return "\"RouterId\":\"" + strconv.Itoa(mgrId) + "?" + strconv.Itoa(clientId) + "\""
}

func AddRoutingForwardRequest(reqMessage string, mgrId int, clientId int, transportMgrChan chan string) {
//	newPrefix := "{ \"RouterId\":\"" + strconv.Itoa(mgrId) + "?" + strconv.Itoa(clientId) + "\", "
	newPrefix := "{" + createRouterIdProperty(mgrId, clientId) + ", "
	request := strings.Replace(reqMessage, "{", newPrefix, 1)
Info.Printf("AddRoutingForwardRequest: %s", request)
	transportMgrChan <- request
}

func backendHttpAppSession(message string, w *http.ResponseWriter) {
	Info.Printf("backendHttpAppSession(): Message received=%s", message)

	var responseMap = make(map[string]interface{})
	MapRequest(message, &responseMap)
	if responseMap["action"] != nil {
		delete(responseMap, "action")
	}
	if responseMap["requestId"] != nil {
		delete(responseMap, "requestId")
	}
	response := FinalizeMessage(responseMap)

	resp := []byte(response)
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
	(*w).Header().Set("Content-Length", strconv.Itoa(len(resp)))
	written, err := (*w).Write(resp)
	if err != nil {
		Error.Printf("HTTP manager error on response write.Written bytes=%d. Error=%s", written, err.Error())
	}
}

func splitToPathQueryKeyValue(path string) (string, string, string) {
	delim := strings.Index(path, "?")
	if delim != -1 {
		if path[delim+1] == 'f' {
			return path[:delim], "filter", path[delim+8:] // path?filter=json-exp
		} else if path[delim+1] == 'm' {
			return path[:delim], "metadata", path[delim+10:] // path?metadata=static (or dynamic)
		}
	}
	return path, "", ""
}

func frontendHttpAppSession(w http.ResponseWriter, req *http.Request, clientChannel chan string) {
	path := req.RequestURI
	if len(path) == 0 {
		path = "empty-path" // will generate error as not found in VSS tree
	}
	path = strings.ReplaceAll(path, "%22", "\"")
	path = strings.ReplaceAll(path, "%20", "")
	var requestMap = make(map[string]interface{})
	queryKey := ""
	queryValue := ""
	if strings.Contains(path, "?") == true {
		requestMap["path"], queryKey, queryValue = splitToPathQueryKeyValue(path)
	} else {
		requestMap["path"] = path
	}
	Info.Printf("HTTP method:%s, path: %s", req.Method, path)
	token := req.Header.Get("Authorization")
	Info.Printf("HTTP token:%s", token)
	if len(token) > 0 {
		requestMap["token"] = token
	}
	requestMap["requestId"] = strconv.Itoa(requestTag)
	requestTag++
	switch req.Method {
	case "OPTIONS":
		fallthrough // should work for POST also...
	case "GET":
		requestMap["action"] = "get"
	case "POST": // set
		requestMap["action"] = "set"
		body, _ := ioutil.ReadAll(req.Body)
		requestMap["value"] = string(body)
	default:
		//		http.Error(w, "400 Unsupported method", http.StatusBadRequest)
		Warning.Printf("Only GET and POST methods are supported.")
		backendHttpAppSession(`{"error": "400", "reason": "Bad request", "message":"Unsupported HTTP method"}`, &w)
		return
	}
	clientChannel <- AddKeyValue(FinalizeMessage(requestMap), queryKey, queryValue) // forward to mgr hub,
	response := <-clientChannel                                                     //  and wait for response

	backendHttpAppSession(response, &w)
}

func frontendWSAppSession(conn *websocket.Conn, clientChannel chan string, clientBackendChannel chan string, clientId int, compression Compression) {
	defer conn.Close()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			Error.Printf("App client read error: %s", err)
			clientChannel <- `{"action":"internal-killsubscriptions"}`
			time.Sleep(100 * time.Millisecond) // to allow for outstanding notifications before backend is killed
			clientBackendChannel <- backendTermination
			ReturnWsClientIndex(clientId)
			break
		}

		var payload string
		if (compression == PROPRIETARY) {
		    payload = string(DecompressMessage(msg))
		} else if (compression == PB_LEVEL1 || compression == PB_LEVEL2) {
		    payload = ProtobufToJson(msg, compression)
		} else {
		    payload = string(msg)
		}
		Info.Printf("%s request: %s, len=%d", conn.RemoteAddr(), payload, len(payload))

		clientChannel <- payload    // forward to mgr hub,
		response := <-clientChannel //  and wait for response

		clientBackendChannel <- response
	}
}

func backendWSAppSession(conn *websocket.Conn, clientBackendChannel chan string, compression Compression) {
	defer conn.Close()
	for {
		message := <-clientBackendChannel
		Info.Printf("backendWSAppSession(): Message received=%s", message)
		if (message == backendTermination) {
			Error.Print("App client websocket session error.")
			break
		}
		// Write response/notification back to app client
		var response []byte
	        var messageType int

		if (compression == PROPRIETARY) {
		    response = CompressMessage([]byte(message))
		    messageType = websocket.BinaryMessage
		} else if (compression == PB_LEVEL1 || compression == PB_LEVEL2) {
		    response = []byte(JsonToProtobuf(message, compression))
		    messageType = websocket.BinaryMessage
		} else {
		    response = []byte(message)
		    messageType = websocket.TextMessage
               }
	        err := conn.WriteMessage(messageType, response)
		if err != nil {
			Error.Print("App client write error:", err)
//			break  // likely to be followed by a read error which trigger the termination above.
		}
	}
}

func (httpH HttpChannel) makeappClientHandler(appClientChannel []chan string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Upgrade") == "websocket" {
			http.Error(w, "400 Incorrect port number", http.StatusBadRequest)
			Warning.Printf("Client call to incorrect port number for websocket connection.")
			return
		}
		frontendHttpAppSession(w, req, appClientChannel[0])
	}
}

func (wsH WsChannel) makeappClientHandler(appClientChannel []chan string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		*wsH.clientIndex = getWsClientIndex()
		if (*wsH.clientIndex == -1) {
		    Warning.Printf("WS session not started, too many clients")
		    return
		}
		if req.Header.Get("Upgrade") == "websocket" {
			Info.Printf("we are upgrading to a websocket connection.")
			Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
			var compression Compression
			compression = NONE
			h := http.Header{}
			for _, sub := range websocket.Subprotocols(req) {
			   if sub == "VISSv2" {
			      compression = NONE
			      h.Set("Sec-Websocket-Protocol", sub)
			      break
			   }
			   if sub == "VISSv2prop" {
			      if (InitCompression("../vsspathlist.json") == true) {
			          compression = PROPRIETARY
			          h.Set("Sec-Websocket-Protocol", sub)
			      } else {
			          Error.Printf("Cannot find vsspathlist.json.")
			          compression = NONE // revert back to no compression
			          h.Set("Sec-Websocket-Protocol", "VISSv2")
			      }
			      break
			   }
			   if sub == "VISSv2pbl1" {
			      compression = PB_LEVEL1
			      h.Set("Sec-Websocket-Protocol", sub)
			      break
			   }
			   if sub == "VISSv2pbl2" {
			      if (InitCompression("../vsspathlist.json") == true) {
			          compression = PB_LEVEL2
			          h.Set("Sec-Websocket-Protocol", sub)
			      } else {
			          Error.Printf("Cannot find vsspathlist.json.")
			          compression = PB_LEVEL1 // revert back to level 1
			          h.Set("Sec-Websocket-Protocol", "VISSv2pbl1")
			      }
			      break
			   }
			}
//			*wsH.clientIndex = getWsClientIndex()
			Info.Printf("ClientIndex=%d", *wsH.clientIndex)
			if *wsH.clientIndex >= 0 {
				conn, err := Upgrader.Upgrade(w, req, h)
				if err != nil {
					Error.Print("upgrade error:", err)
					return
				}
				Info.Printf("WS session started, compression variant=%d", compression)
				go frontendWSAppSession(conn, appClientChannel[*wsH.clientIndex], wsH.clientBackendChannel[*wsH.clientIndex], 
							*wsH.clientIndex, compression)
				go backendWSAppSession(conn, wsH.clientBackendChannel[*wsH.clientIndex], compression)
			}
		} else {
			Error.Printf("Client must set up a Websocket session.")
		}
	}
}

func (server HttpServer) InitClientServer(muxServer *http.ServeMux, httpClientChan []chan string) {

	appClientHandler := HttpChannel{}.makeappClientHandler(httpClientChan)
	muxServer.HandleFunc("/", appClientHandler)
	Info.Printf("InitClientServer():secConfig.TransportSec=%s", secConfig.TransportSec)
	if secConfig.TransportSec == "yes" {
		secPortNum, _ := strconv.Atoi(secConfig.HttpSecPort)
		server := http.Server{
			Addr: ":" + strconv.Itoa(secPortNum),
			TLSConfig: getTLSConfig("localhost", trSecConfigPath+secConfig.CaSecPath+"Root.CA.crt",
				tls.ClientAuthType(certOptToInt(secConfig.ServerCertOpt))),
			Handler: muxServer,
		}
		Info.Printf("HTTPS:CerOpt=%s", secConfig.ServerCertOpt)
		Error.Fatal(server.ListenAndServeTLS(trSecConfigPath+secConfig.ServerSecPath+"server.crt", trSecConfigPath+secConfig.ServerSecPath+"server.key"))
	} else {
		Error.Fatal(http.ListenAndServe(":8888", muxServer))
	}
}

func (server WsServer) InitClientServer(muxServer *http.ServeMux, wsClientChan []chan string, mgrIndex int, clientIndex *int) {
	*clientIndex = 0
	appClientHandler := WsChannel{server.ClientBackendChannel, mgrIndex, clientIndex}.makeappClientHandler(wsClientChan)
	muxServer.HandleFunc("/", appClientHandler)
	Info.Printf("InitClientServer():secConfig.TransportSec=%s", secConfig.TransportSec)
	if secConfig.TransportSec == "yes" {
		server := http.Server{
			Addr: ":" + secConfig.WsSecPort,
			TLSConfig: getTLSConfig("localhost", trSecConfigPath+secConfig.CaSecPath+"Root.CA.crt",
				tls.ClientAuthType(certOptToInt(secConfig.ServerCertOpt))),
			Handler: muxServer,
		}
		Info.Printf("HTTPS:CerOpt=%s", secConfig.ServerCertOpt)
		Error.Fatal(server.ListenAndServeTLS(trSecConfigPath+secConfig.ServerSecPath+"server.crt", trSecConfigPath+secConfig.ServerSecPath+"server.key"))
	} else {
		Error.Fatal(http.ListenAndServe(":8080", muxServer))
	}
}

func certOptToInt(serverCertOpt string) int {
	if serverCertOpt == "NoClientCert" {
		return 0
	}
	if serverCertOpt == "ClientCertNoVerification" {
		return 2
	}
	if serverCertOpt == "ClientCertVerification" {
		return 4
	}
	return 4 // if unclear, apply max security
}

func getTLSConfig(host string, caCertFile string, certOpt tls.ClientAuthType) *tls.Config {
	var caCert []byte
	var err error
	var caCertPool *x509.CertPool
	if certOpt > tls.RequestClientCert {
		caCert, err = ioutil.ReadFile(caCertFile)
		if err != nil {
			Error.Printf("Error opening cert file", caCertFile, ", error ", err)
			return nil
		}
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}

	return &tls.Config{
		ServerName: host,
		ClientAuth: certOpt,
		ClientCAs:  caCertPool,
		MinVersion: tls.VersionTLS12, // TLS versions below 1.2 are considered insecure - see https://www.rfc-editor.org/rfc/rfc7525.txt for details
	}
}

func RemoveInternalData(response string) (string, int) { // "RouterId":"mgrId?clientId",
	routerIdStart := strings.Index(response, "RouterId") - 1
	clientIdStart := strings.Index(response[routerIdStart:], "?") + 1 + routerIdStart
	clientIdStop := NextQuoteMark([]byte(response), clientIdStart)
	clientId, _ := strconv.Atoi(response[clientIdStart:clientIdStop])
	routerIdStop := strings.Index(response[clientIdStop:], ",") + 1 + clientIdStop
	trimmedResponse := response[:routerIdStart] + response[routerIdStop:]
//Info.Printf("response=%s, trimmedResponse=%s, clientId=%d", response, trimmedResponse, clientId)
	return trimmedResponse, clientId
}

