package main

import (
	"log"
	"os"
	"strconv"

	"github.com/Diverstt/Calculator_Yandex/internal/agent"
)

func main() {
	workers, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if err != nil || workers <= 0 {
		workers = 2
	}
	log.Printf("Запуск агента с %d рабочими горутинами...", workers)
	agent.StartWorkers(workers)
	select {}
}
