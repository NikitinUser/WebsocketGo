package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/NikitinUser/WebsocketGo/pkg/connect_storage"

	"github.com/wagslane/go-rabbitmq"
)

var consumeModes = map[string]interface{}{
	"all":    sendToAll,
	"touser": sendToUser,
}

func Consume() {
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

	ipPorts, exist := connect_storage.Users[userid]
	if !exist {
		return
	}

	for _, ipPort := range ipPorts {
		client, exist := connect_storage.Connections[ipPort]
		if !exist {
			continue
		}

		client.Connection.WriteMessage(1, []byte(message))
	}
}

func sendToAll(unserializedMsg map[string]interface{}) {
	message, okMessage := unserializedMsg["message"].(string)
	if !okMessage {
		return
	}

	for _, ipPorts := range connect_storage.Users {
		for _, ipPort := range ipPorts {
			client, exist := connect_storage.Connections[ipPort]
			if !exist {
				continue
			}

			client.Connection.WriteMessage(1, []byte(message))
		}
	}
}
