package user_service_sdk

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type Ticket struct {
	Ipaddr string `json:"ipaddr"`
	Userid string `json:"userid"`
}

func GetTicketData(t string) Ticket {
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
