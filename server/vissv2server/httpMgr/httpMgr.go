/**
* (C) 2022 Geotab
* (C) 2019 Volvo Cars
*
* All files and artifacts in the repository at https://github.com/w3c/automotive-viss2
* are licensed under the provisions of the license provided by the LICENSE file in this repository.
*
**/
package httpMgr

import (
	"github.com/w3c/automotive-viss2/utils"
)

// All HTTP app clients share same channel
var HttpClientChan = []chan string{
	make(chan string),
}

func RemoveRoutingForwardResponse(response string, transportMgrChan chan string) {
	trimmedResponse, clientId := utils.RemoveInternalData(response)
	HttpClientChan[clientId] <- trimmedResponse
}

func HttpMgrInit(mgrId int, transportMgrChan chan string, trSecConfigPath string) {
//	utils.ReadTransportSecConfig(trSecConfigPath)
	utils.ReadTransportSecConfig()

	go utils.HttpServer{}.InitClientServer(utils.MuxServer[0], HttpClientChan) // go routine needed due to listenAndServe call...
	utils.Info.Println("HTTP manager data session initiated.")

	utils.Info.Println("**** HTTP manager entering server loop... ****")
	for {
		select {
		case reqMessage := <-HttpClientChan[0]:
			utils.Info.Printf("HTTP mgr hub: Request from client:%s\n", reqMessage)
			utils.AddRoutingForwardRequest(reqMessage, mgrId, 0, transportMgrChan)
		case respMessage := <-transportMgrChan:
			utils.Info.Printf("HTTP mgr hub: Response from server core:%s\n", respMessage)
			RemoveRoutingForwardResponse(respMessage, transportMgrChan)
		}
	}
}
