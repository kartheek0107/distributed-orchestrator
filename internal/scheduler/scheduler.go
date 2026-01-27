package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	pb "github.com/kartheek0107/distributed-orchestrator/api/proto"
	"github.com/kartheek0107/distributed-orchestrator/pkg/models"
)

type scheduler struct {
	pb.UnimplementedCoordinatorServiceServer

	mu      sync.RWMutex
	workers map[string]time.Time
	jobs    map[string]*models.Job
}

func NewScheduler() *scheduler {
	return &scheduler{
		workers: make(map[string]time.Time),
		jobs:    make(map[string]*models.Job),
	}
}

func (s *scheduler) AddJob(job *models.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	log.Printf("Job %s added to the scheduler", job.ID)
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

func (s *scheduler) StartMonitor(ctx context.Context, interval time.Duration, timeout time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkWorkerHealth(timeout)
		}
	}
}

func (s *scheduler) checkWorkerHealth(timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, lastseen := range s.workers {
		if now.Sub(lastseen) > timeout {
			log.Printf("Worker %s timed out! Removing from registry.", id)
			delete(s.workers, id)
		}
	}
}
