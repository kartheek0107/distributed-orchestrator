package scheduler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kartheek0107/distributed-orchestrator/pkg/models"
)

type JobHandler struct {
	scheduler *scheduler
}

func NewJobHandler(s *scheduler) *JobHandler {
	return &JobHandler{
		scheduler: s,
	}
}

func (h *JobHandler) SubmitJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var job models.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	job.ID = uuid.New().String()
	job.Status = "pending"
	job.CreatedAt = time.Now().Unix()

	h.scheduler.AddJob(&job)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"job_id": job.ID})
}
