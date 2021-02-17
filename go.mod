module github.com/MEAE-GOT/W3C_VehicleSignalInterfaceImpl

go 1.13

//example on how to use replace to point to fork or local path
//replace github.com/MEAE-GOT/W3C_VehicleSignalInterfaceImpl/utils => github.com/MagnusGun/W3C_VehicleSignalInterfaceImpl/utils master

require (
	github.com/eclipse/paho.mqtt.golang v1.3.2
	github.com/golang/protobuf v1.4.1
	github.com/gorilla/websocket v1.4.2
	github.com/magefile/mage v1.11.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.3
	github.com/sirupsen/logrus v1.7.1
	golang.org/x/net v0.0.0-20200505041828-1ed23360d12c // indirect
	golang.org/x/sys v0.0.0-20210217105451-b926d437f341 // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20200430143042-b979b6f78d84 // indirect
	google.golang.org/grpc v1.29.1
)
