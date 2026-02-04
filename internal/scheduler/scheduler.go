package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	pb "github.com/kartheek0107/distributed-orchestrator/api/proto"
	"github.com/kartheek0107/distributed-orchestrator/internal/queue"
	"github.com/kartheek0107/distributed-orchestrator/pkg/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type scheduler struct {
	pb.UnimplementedCoordinatorServiceServer

	mu      sync.RWMutex
	workers map[string]time.Time
	jobs    map[string]*models.Job
	queue   *queue.RedisQueue
}

func NewScheduler(q *queue.RedisQueue) *scheduler {
	return &scheduler{
		workers: make(map[string]time.Time),
		jobs:    make(map[string]*models.Job),
		queue:   q, // Initialize with actual RedisQueue instance
	}
}

func (s *scheduler) AddJob(job *models.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	err := s.queue.Enqueue(context.Background(), job)
	if err != nil {
		log.Printf("Failed to enqueue job %s: %v", job.ID, err)
		return
	}
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

	log.Printf("Final Result %s completed by %s | success : %v", req.JobId, req.WorkerId, req.Success)
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

func (s *scheduler) StartDistributor(ctx context.Context) {
	log.Println("Distributor started, waiting for jobs from Redis...")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 1. Dequeue blocks until a job is available
			job, err := s.queue.Dequeue(ctx)
			if err != nil {
				log.Printf("❌ Failed to dequeue job: %v", err)
				continue
			}
			if job == nil {
				continue
			}

			// 2. Try to find a worker
			workerID := s.FindWorker()
			if workerID == "" {
				log.Printf("⚠️ No available workers for job %s. Re-enqueuing...", job.ID)
				s.AddJob(job)
				time.Sleep(2 * time.Second) // Wait longer so we don't spam Redis
				continue
			}

			log.Printf("📡 Attempting to push job %s to worker %s", job.ID, workerID)

			// 3. Connect to Worker (Modern gRPC Dial)
			// Ensure worker is actually on 50052!

			opts := []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			}

			conn, err := grpc.Dial("localhost:50052", opts...)
			if err != nil {
				log.Printf("❌ Connection error to worker %s: %v", workerID, err)
				continue
			}

			workerClient := pb.NewWorkerServiceClient(conn)

			// 4. Send the task
			resp, err := workerClient.StartTask(ctx, &pb.TaskRequest{
				JobId:   job.ID,
				Command: job.Command,
			})

			if err != nil {
				log.Printf("❌ gRPC StartTask Failed for worker %s: %v", workerID, err)
			} else {
				log.Printf("🚀 SUCCESS: Worker %s accepted task: %s", workerID, resp.Message)
			}

			conn.Close()
		}
	}
}

func (s *scheduler) FindWorker() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id := range s.workers {
		return id
	}
	return ""
}
