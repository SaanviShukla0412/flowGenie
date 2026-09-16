package main

import (
	"context"
	"log"

	"github.com/SaanviShukla0412/flowGenie/internal/database"
	"github.com/SaanviShukla0412/flowGenie/internal/executor"
	"github.com/SaanviShukla0412/flowGenie/internal/queue"
	"github.com/SaanviShukla0412/flowGenie/internal/worker"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer db.Close(context.Background())

	log.Println("Connected to PostgreSQL")
	redisQueue := queue.NewRedisQueue()

	if err := redisQueue.Ping(context.Background()); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	log.Println("Connected to Redis")

	stepExecutor := executor.NewExecutor()

	w := worker.NewWorker(redisQueue, stepExecutor, db)

	w.Start(context.Background())
}
