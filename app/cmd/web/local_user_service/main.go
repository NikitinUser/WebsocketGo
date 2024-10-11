package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/check/ticket", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"ipaddr\": \"127.0.0.1\", \"userid\": \"1\"}")
	})

	log.Println("192.168.0.189:6070/check/ticket")
	log.Fatal(http.ListenAndServe(":6070", nil))
}
