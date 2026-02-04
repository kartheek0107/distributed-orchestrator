package worker

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"time"

	pb "github.com/kartheek0107/distributed-orchestrator/api/proto"
)

// Worker represents the execution node and implements the WorkerService gRPC interface
type Worker struct {
	pb.UnimplementedWorkerServiceServer
	ID                string
	SchedulerClient   pb.CoordinatorServiceClient
	HeartbeatInterval time.Duration
}

func NewWorker(id string, client pb.CoordinatorServiceClient, interval time.Duration) *Worker {
	return &Worker{
		ID:                id,
		SchedulerClient:   client,
		HeartbeatInterval: interval,
	}
}

// StartHeartbeat begins the periodic heartbeat loop
func (w *Worker) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(w.HeartbeatInterval)
	defer ticker.Stop()

	log.Printf("Worker %s starting heartbeat loop (interval: %v)", w.ID, w.HeartbeatInterval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %s stopping heartbeat loop", w.ID)
			return
		case <-ticker.C:
			w.performHeartbeat()
		}
	}
}

func (w *Worker) performHeartbeat() {
	req := &pb.HeartbeatRequest{
		WorkerId:  w.ID,
		Timestamp: time.Now().Unix(),
		// Later we can add ActiveTasks count here
	}

	// Use a short timeout for the RPC call
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := w.SchedulerClient.ReportHeartbeat(ctx, req)
	if err != nil {
		log.Printf("Failed to send heartbeat for %s: %v", w.ID, err)
		return
	}
	log.Printf("Heartbeat sent for worker %s", w.ID)
}

// StartTask is called by the Scheduler via gRPC to assign work
func (w *Worker) StartTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	log.Printf("Worker %s received task: %s (Command: %s)", w.ID, req.JobId, req.Command)

	// Execute task in background so the gRPC call returns quickly
	go w.executeTask(req.JobId, req.Command)

	return &pb.TaskResponse{
		Accepted: true,
		Message:  "Task accepted " + req.JobId + " for execution",
	}, nil
}

func (w *Worker) executeTask(taskID string, command string) {
	log.Printf("[Worker %s] Executing: %s", w.ID, command)

	// 1. Run the actual command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()

	// 2. Prepare the completion report
	success := (err == nil)
	resultMsg := string(output)
	if err != nil {
		resultMsg = err.Error()
	}

	// 3. Call Scheduler via gRPC
	req := &pb.CompletionRequest{
		JobId:    taskID,
		WorkerId: w.ID,
		Success:  success,
		Result:   resultMsg,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, reportErr := w.SchedulerClient.ReportCompletion(ctx, req)
	if reportErr != nil {
		log.Printf("Failed to report completion for %s: %v", taskID, reportErr)
		return
	}

	log.Printf("Successfully reported completion for task %s (Success: %v)", taskID, success)
}
