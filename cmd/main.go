package main

import (
	"fmt"
	"log"

	"octopus/infostructure/logger"
	"octopus/infostructure/security"
	"octopus/infostructure/server"
)

func main() {
	fmt.Println("http://localhost:8080")
	fmt.Println("http://localhost:8080/logs")

	logSrv := logger.New("localhost:8080", "log.txt")
	go func() {
		if err := logSrv.Run(); err != nil {
			log.Fatalf("logger server failed: %v", err)
		}
	}()

	if err := server.Run("localhost", 1488, "http://localhost:8080/log"); err != nil {
		log.Fatalf("main server failed: %v", err)
	}

	secGuard := security.New(5)
	_ = secGuard
}
