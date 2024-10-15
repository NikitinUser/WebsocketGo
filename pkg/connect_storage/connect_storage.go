package connect_storage

import (
	"slices"

	"github.com/gorilla/websocket"
)

type Connect struct {
	Userid     string
	Connection *websocket.Conn
}

var Connections = map[string]*Connect{}
var Users = map[string][]string{}

func SaveConnection(clientIP string, port string, userid string, conn *websocket.Conn) string {
	ipPort := clientIP + "_" + port
	Connections[ipPort] = &Connect{Userid: userid, Connection: conn}
	Users[userid] = append(Users[userid], ipPort)

	return ipPort
}

func DeleteConn(ipPort string) {
	userid := Connections[ipPort].Userid

	Users[userid] = slices.DeleteFunc(Users[userid], func(cmp string) bool {
		return cmp == ipPort
	})

	if len(Users[userid]) == 0 {
		delete(Users, userid)
	}

	delete(Connections, ipPort)
}
