package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func startEchoServer(port int) {
	addr := fmt.Sprintf(":%d", port)

	// 독립적인 ServeMux 생성
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // 2초 지연
		fmt.Fprintf(w, "Hello from port %d\n", port)
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("Echo server listening on %s", addr)
	log.Fatal(server.ListenAndServe())
}
func main() {
	ports := []int{8081, 8082, 8083, 8084, 8085}

	for _, port := range ports {
		go startEchoServer(port)
	}

	// main goroutine이 끝나지 않도록 대기
	select {}
}
