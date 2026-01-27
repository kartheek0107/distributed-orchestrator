package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	pb "github.com/kartheek0107/distributed-orchestrator/api/proto"
	"github.com/kartheek0107/distributed-orchestrator/internal/scheduler"
	"google.golang.org/grpc"
)

func main() {
	// TODO: Implement scheduler

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcserver := grpc.NewServer()

	sched := scheduler.NewScheduler()

	handler := scheduler.NewJobHandler(sched)

	go func() {
		http.HandleFunc("/api/v1/jobs", handler.SubmitJob)
		log.Printf("REST API listening on 8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("failed to start REST API server: %v", err)
		}
	}()

	go sched.StartMonitor(context.Background(), 5*time.Second, 10*time.Second)

	pb.RegisterCoordinatorServiceServer(grpcserver, sched)
	if err := grpcserver.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	log.Printf("Scheduler is listening on %v", lis.Addr())

}
