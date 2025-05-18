package main

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"

	"github.com/meesooqa/files2tg/app/job"
	"github.com/meesooqa/files2tg/app/send"
	"github.com/meesooqa/files2tg/app/web"
)

func main() {
	godotenv.Load()

	tgFactory := &send.EnvClientFactory{}
	tgClient, err := tgFactory.NewClient()
	if err != nil {
		fmt.Printf("new tgClient: %v\n", err)
		return
	}

	jq := job.NewJobQueue()
	// Start workers
	numWorkers := 1
	for i := 1; i <= numWorkers; i++ {
		go job.Worker(i, jq)
	}
	server := web.Server{
		JobQueue:       jq,
		TelegramClient: tgClient,
	}
	server.Run(context.Background(), 8080)
}
