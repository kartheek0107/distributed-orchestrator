package scheduler

import (
	"context"
	"sync"
	"time"

	pb "github.com/kartheek0107/distributed-orchestrator/api/proto"
)

type scheduler struct {
	pb.UnimplementedCoordinatorServiceServer

	mu      sync.RWMutex
	workers map[string]time.Time
}

func NewScheduler() *scheduler {
	return &scheduler{
		workers: make(map[string]time.Time),
	}
}

func (s *scheduler) ReportHeartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	s.workers[req.WorkerId] = time.Now()

	return &pb.HeartbeatResponse{Acknowledged: true}, nil
}

func (s *scheduler) SubmitJob(ctx context.Context, req *pb.JobRequest) (*pb.JobResponse, error) {
	// Job submission logic would go here
	return &pb.JobResponse{
		JobId:   "job-12345",
		Success: true}, nil
}

func (s *scheduler) ReportCompletion(ctx context.Context, req *pb.CompletionRequest) (*pb.CompletionResponse, error) {
	// Job completion handling logic would go here
	return &pb.CompletionResponse{
		Acknowledged: true}, nil
}
