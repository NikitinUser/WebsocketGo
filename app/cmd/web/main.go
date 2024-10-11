package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/wagslane/go-rabbitmq"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Ticket struct {
	Ipaddr string `json:"ipaddr"`
	Userid string `json:"userid"`
}

type connect struct {
	userid     string
	connection *websocket.Conn
}

var connections = map[string]*connect{}
var users = map[string][]string{}

var consumeModes = map[string]interface{}{
	"all":    sendToAll,
	"touser": sendToUser,
}

func main() {
	envPath := filepath.Join(".env")
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("env err load: %v", err)
	}

	websocketPort := os.Getenv("WEBSOCKET_PORT")

	go rabbitConsumer()

	http.HandleFunc("/ws", wsHandler)
	log.Printf("Run in :%s", websocketPort)
	log.Fatal(http.ListenAndServe(":"+websocketPort, nil))
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

	ticketData := getTicketData(ticket)
	userid := ticketData.Userid
	expectedIP := ticketData.Ipaddr

	if os.Getenv("APP_ENV") == "prod" && clientIP != expectedIP {
		http.Error(w, "IP адрес не совпадает", http.StatusForbidden)
		return
	}

	ipPort := saveConnection(clientIP, port, userid, conn)

	go pingPongHandler(conn, ipPort)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Клиент отключился: %v", err)
			deleteConn(ipPort)
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
				deleteConn(ipPort)
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

func getTicketData(t string) Ticket {
	url := os.Getenv("USER_SERVICE_HOST") + "?ticket=" + t
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("USER_SERVICE_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var ticket Ticket

	err = json.Unmarshal(body, &ticket)
	if err != nil {
		log.Fatal(err)
	}

	return ticket
}

func saveConnection(clientIP string, port string, userid string, conn *websocket.Conn) string {
	ipPort := clientIP + "_" + port
	connections[ipPort] = &connect{userid: userid, connection: conn}
	users[userid] = append(users[userid], ipPort)

	return ipPort
}

func deleteConn(ipPort string) {
	userid := connections[ipPort].userid

	users[userid] = slices.DeleteFunc(users[userid], func(cmp string) bool {
		return cmp == ipPort
	})

	if len(users[userid]) == 0 {
		delete(users, userid)
	}

	delete(connections, ipPort)
}

func rabbitConsumer() {
	rabbitMQHost := os.Getenv("RABBITMQ_HOST")
	rabbitMQUser := os.Getenv("RABBITMQ_USER")
	rabbitMQPassword := os.Getenv("RABBITMQ_PASSWORD")
	rabbitMQVHost := os.Getenv("RABBITMQ_VHOST")
	queue := os.Getenv("OUTPUT_QUEUE")

	rabbitMQURL := fmt.Sprintf("amqp://%s:%s@%s/%s",
		rabbitMQUser, rabbitMQPassword, rabbitMQHost, rabbitMQVHost)

	conn, err := rabbitmq.NewConn(
		rabbitMQURL,
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		log.Fatalf("Ошибка плдключения: %v", err)
	}
	defer conn.Close()

	consumer, err := rabbitmq.NewConsumer(
		conn,
		queue,
	)
	if err != nil {
		log.Fatalf("Ошибка создания потребителя: %v", err)
	}

	err = consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
		outputHandler(d.Body)
		return rabbitmq.Ack
	})
	if err != nil {
		log.Fatal(err)
	}
}

func outputHandler(msg []byte) {
	log.Printf("consumed: %v", string(msg))

	var unserializedMsg map[string]interface{}

	err := json.Unmarshal(msg, &unserializedMsg)
	if err != nil {
		log.Println(err)
		return
	}

	mode, okMode := unserializedMsg["mode"].(string)
	if !okMode {
		return
	}

	strategy, exist := consumeModes[mode]
	if exist {
		strategy.(func(map[string]interface{}))(unserializedMsg)
	}
}

func sendToUser(unserializedMsg map[string]interface{}) {
	message, okMessage := unserializedMsg["message"].(string)
	if !okMessage {
		return
	}

	userid, okUserid := unserializedMsg["userid"].(string)
	if !okUserid {
		return
	}

	ipPorts, exist := users[userid]
	if !exist {
		return
	}

	for _, ipPort := range ipPorts {
		client, exist := connections[ipPort]
		if !exist {
			continue
		}

		client.connection.WriteMessage(1, []byte(message))
	}
}

func sendToAll(unserializedMsg map[string]interface{}) {
	message, okMessage := unserializedMsg["message"].(string)
	if !okMessage {
		return
	}

	for _, ipPorts := range users {
		for _, ipPort := range ipPorts {
			client, exist := connections[ipPort]
			if !exist {
				continue
			}

			client.connection.WriteMessage(1, []byte(message))
		}
	}
}
