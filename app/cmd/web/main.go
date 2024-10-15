package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	"main/pkg/connect_storage"
	"main/pkg/user_service_sdk"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	envPath := filepath.Join(".env")
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("env err load: %v", err)
	}

	go Consume()

	http.HandleFunc("/ws", wsHandler)
	log.Printf("Run in :%s", os.Getenv("WEBSOCKET_PORT"))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("WEBSOCKET_PORT"), nil))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	clientIP, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Невозможно определить IP клиента", http.StatusInternalServerError)
		return
	}

	ticket := r.URL.Query().Get("ticket")
	if ticket == "" {
		http.Error(w, "Ticket не указан", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка при апгрейде соединения:", err)
		return
	}
	defer conn.Close()

	ticketData := user_service_sdk.GetTicketData(ticket)
	userid := ticketData.Userid
	expectedIP := ticketData.Ipaddr

	if os.Getenv("APP_ENV") == "prod" && clientIP != expectedIP {
		http.Error(w, "IP адрес не совпадает", http.StatusForbidden)
		return
	}

	ipPort := connect_storage.SaveConnection(clientIP, port, userid, conn)

	go pingPongHandler(conn, ipPort)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Клиент отключился: %v", err)
			connect_storage.DeleteConn(ipPort)
			break
		}
	}
}

func pingPongHandler(conn *websocket.Conn, ipPort string) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("Ошибка отправки пинга:", err)
				connect_storage.DeleteConn(ipPort)
				conn.Close()
				return
			}
			log.Println("Пинг отправлен")

			conn.SetPongHandler(func(appData string) error {
				log.Println("Получен понг")
				return nil
			})
		}
	}
}
