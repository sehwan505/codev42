package queue

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// Job represents an implementation job
type Job struct {
	ID           string
	DevPlanID    int64
	Status       JobStatus
	Progress     int32
	CurrentStep  string
	Result       *JobResult
	Error        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
}

// JobResult stores the implementation result
type JobResult struct {
	Code              string
	Diagrams          []Diagram
	ExplainedSegments []ExplainedSegment
}

// Diagram represents a mermaid diagram
type Diagram struct {
	Diagram string
	Type    string
}

// ExplainedSegment represents an explained code segment
type ExplainedSegment struct {
	StartLine   int32
	EndLine     int32
	Explanation string
}

// JobQueue is an in-memory job queue (can be replaced with Redis)
type JobQueue struct {
	jobs map[string]*Job
	mu   sync.RWMutex
}

// NewJobQueue creates a new job queue
func NewJobQueue() *JobQueue {
	return &JobQueue{
		jobs: make(map[string]*Job),
	}
}

// CreateJob creates a new job
func (q *JobQueue) CreateJob(devPlanID int64) *Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	job := &Job{
		ID:          uuid.New().String(),
		DevPlanID:   devPlanID,
		Status:      JobStatusPending,
		Progress:    0,
		CurrentStep: "Initializing",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	q.jobs[job.ID] = job
	return job
}

// GetJob retrieves a job by ID
func (q *JobQueue) GetJob(jobID string) (*Job, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	job, exists := q.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// UpdateJob updates a job's status and progress
func (q *JobQueue) UpdateJob(jobID string, status JobStatus, progress int32, currentStep string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	job, exists := q.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Status = status
	job.Progress = progress
	job.CurrentStep = currentStep
	job.UpdatedAt = time.Now()

	if status == JobStatusCompleted || status == JobStatusFailed {
		now := time.Now()
		job.CompletedAt = &now
	}

	return nil
}

// SetJobResult sets the result of a completed job
func (q *JobQueue) SetJobResult(jobID string, result *JobResult) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	job, exists := q.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Result = result
	return nil
}

// SetJobError sets the error for a failed job
func (q *JobQueue) SetJobError(jobID string, err error) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	job, exists := q.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Error = err.Error()
	job.Status = JobStatusFailed
	now := time.Now()
	job.CompletedAt = &now
	job.UpdatedAt = time.Now()

	return nil
}
