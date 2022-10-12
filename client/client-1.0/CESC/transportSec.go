/**
* (C) 2021 Geotab Inc
*
* All files and artifacts in the repository at https://github.com/GENIVI/ccs-w3c-client
* are licensed under the provisions of the license provided by the LICENSE file in this repository.
*
**/

package main

import (
	"encoding/json"
	"io/ioutil"
	"fmt"
//	"os"
//	"crypto/tls"
//	"crypto/x509"
)


var trSecConfigPath string = "transport_sec/"  // relative path to the directory containing the transportSec.json file
type SecConfig struct {
    TransportSec string  `json:"transportSec"`// "yes" or "no"
    HttpSecPort string   `json:"httpSecPort"`// HTTPS port number
    WsSecPort string     `json:"wsSecPort"`// WSS port number
    MqttSecPort string   `json:"mqttSecPort"`// MQTTS port number
    AgtsSecPort string   `json:"agtsSecPort"`// AGTS port number
    AtsSecPort string    `json:"atsSecPort"`// ATS port number
    CaSecPath string     `json:"caSecPath"`// relative path from the directory containing the transportSec.json file
    ServerSecPath string `json:"serverSecPath"`// relative path from the directory containing the transportSec.json file
    ServerCertOpt string `json:"serverCertOpt"`// one of  "NoClientCert"/"ClientCertNoVerification"/"ClientCertVerification"
    ClientSecPath string `json:"clientSecPath"`// relative path from the directory containing the transportSec.json file
}
var secConfig SecConfig

func readTransportSecConfig() {
	data, err := ioutil.ReadFile(trSecConfigPath + "transportSec.json")
	if err != nil {
	    fmt.Printf("ReadTransportSecConfig():%stransportSec.json error=%s", trSecConfigPath, err)
	    secConfig.TransportSec = "no"
	    return
	}
	err = json.Unmarshal(data, &secConfig)
	if err != nil {
	    fmt.Printf("ReadTransportSecConfig():Error unmarshal transportSec.json=%s", err)
	    secConfig.TransportSec = "no"
	    return
	}
        fmt.Printf("ReadTransportSecConfig():secConfig.TransportSec=%s", secConfig.TransportSec)
}

