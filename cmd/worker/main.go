package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "github.com/kartheek0107/distributed-orchestrator/api/proto"
	"github.com/kartheek0107/distributed-orchestrator/internal/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 1. Establish a connection to the Scheduler (Coordinator)
	// We use port 50051 where the Scheduler is listening
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to scheduler: %v", err)
	}
	defer conn.Close()

	ln, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen on :50052: %v", err)
	}

	// 2. Create the gRPC client for the Coordinator service
	schedulerClient := pb.NewCoordinatorServiceClient(conn)

	// 3. Initialize the Worker with the REAL client (not nil)
	w := worker.NewWorker("worker-1", schedulerClient, 3*time.Second)

	// 4. Start the heartbeat loop in a separate goroutine
	go w.StartHeartbeat(context.Background())

	// 5. Start the Worker's OWN gRPC server to receive tasks from the Scheduler

	grpcServer := grpc.NewServer()

	// We register 'w' because we implemented the WorkerService logic
	// (like StartTask) inside the internal/worker package.
	pb.RegisterWorkerServiceServer(grpcServer, w)

	log.Println("Worker server listening on :50052")
	if err := grpcServer.Serve(ln); err != nil {
		log.Fatalf("Failed to serve worker gRPC: %v", err)
	}
	
}
