package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ywmei-brt1/api101/internal/handlers"
)

// Start the server with: go run main.go
// curl -i -w "\nExit Code: %{http_code} " -X PUT -d "hello" http://localhost:8080/put
// curl -i -w "\nExit Code: %{http_code} " http://localhost:8080/get
// curl -i -w "\nExit Code: %{http_code} " http://localhost:8080/get/longpoll
// curl -i -w "\nExit Code: %{http_code} " "http://localhost:8080/search?q=hello"
func main() {
	http.HandleFunc("/put", handlers.PutHandler)
	http.HandleFunc("/get", handlers.GetHandler)
	http.HandleFunc("/search", handlers.SearchHandler)
	http.HandleFunc("/get/longpoll", handlers.LongPollHandler)

	fmt.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
