package main

import (
	"fmt"
	"net/http"
)

// HelloWorld — обработчик запроса.
func HelloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Println("<h1>Hello, World</h1>")
}

func main() {
	// маршрутизация запросов обработчику
	http.HandleFunc("/hel", HelloWorld)
	// конструируем свой сервер
	server := &http.Server{
		Addr: "localhost:8080",
	}
	server.ListenAndServe()
	fmt.Println("we a started")
}
