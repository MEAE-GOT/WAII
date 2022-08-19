module github.com/w3c/automotive-viss2

go 1.16

//example on how to use replace to point to fork or local path
//replace github.com/w3c/automotive-viss2/utils => github.com/MagnusGun/WAII/utils master
replace github.com/w3c/automotive-viss2/utils => ./utils

replace (
	github.com/COVESA/vss-tools/binary/go_parser/datamodel => github.com/UlfBj/vss-tools/binary/go_parser/datamodel v0.0.0-20220524163944-c753a539973f
	github.com/COVESA/vss-tools/binary/go_parser/parserlib => github.com/UlfBj/vss-tools/binary/go_parser/parserlib v0.0.0-20220524163944-c753a539973f
	github.com/w3c/automotive-viss2/server/vissv2server/atServer => ./server/vissv2server/atServer
	github.com/w3c/automotive-viss2/server/vissv2server/httpMgr => ./server/vissv2server/httpMgr
	github.com/w3c/automotive-viss2/server/vissv2server/mqttMgr => ./server/vissv2server/mqttMgr
	github.com/w3c/automotive-viss2/server/vissv2server/serviceMgr => ./server/vissv2server/serviceMgr
	github.com/w3c/automotive-viss2/server/vissv2server/wsMgr => ./server/vissv2server/wsMgr
	vissv2-server/ecuFeeder => ./server/vissv2server/ecuFeeder
)

//replace github.com/w3c/automotive-viss2/protobuf/protoc-out => ./protobuf/protoc-out

require (
	github.com/COVESA/vss-tools/binary/go_parser/datamodel v0.0.0-20220104185813-cad8492de65f
	github.com/COVESA/vss-tools/binary/go_parser/parserlib v0.0.0-20220104185813-cad8492de65f
	github.com/akamensky/argparse v1.3.1
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/websocket v1.4.2
	github.com/mattn/go-sqlite3 v1.14.14
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/net v0.0.0-20211020060615-d418f374d309 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/protobuf v1.27.1
	modernc.org/sqlite v1.18.1
	vissv2-server/ecuFeeder v0.0.0-00010101000000-000000000000
)
