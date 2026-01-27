package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "github.com/kartheek0107/distributed-orchestrator/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	addr string
}

func NewServer(addr string) *server {
	return &server{addr: addr}
}

func (s *server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	return nil
}

type WorkerServer struct {
	pb.UnimplementedWorkerServiceServer
}

func (s *WorkerServer) StartTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	log.Printf("Received task: ID=%v, Command=%v", req.JobId, req.Command)
	return &pb.TaskResponse{
		Accepted: true,
		Message:  "Task started",
	}, nil
}

func sendHeartbeat(client pb.CoordinatorServiceClient) {
	ticker := time.NewTicker(3 * time.Second)
	for {
		<-ticker.C

		req := &pb.HeartbeatRequest{
			WorkerId:  "worker-1",
			Timestamp: time.Now().Unix(),
		}
		_, err := client.ReportHeartbeat(context.Background(), req)
		if err != nil {
			log.Printf("Failed to send heartbeat: %v", err)
		} else {
			log.Println("Heartbeat sent")
		}
	}
}

func main() {
	// TODO: Implement worker
	ln, err := net.Listen("tcp", ":50052")

	if err != nil {
		log.Fatal("Failed to listen: %v", err)
	}

	grpcserver := grpc.NewServer()
	pb.RegisterWorkerServiceServer(grpcserver, &WorkerServer{})

	log.Println("Worker server listening on :50052")

	grpcclient, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	pb.NewCoordinatorServiceClient(grpcclient)

	go sendHeartbeat(pb.NewCoordinatorServiceClient(grpcclient))
	if err := grpcserver.Serve(ln); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
