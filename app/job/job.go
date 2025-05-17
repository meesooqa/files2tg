package job

import (
	"log"
	"sync"
)

// JobStatus describes the possible statuses of a job
type JobStatus string

const (
	StatusQueued     JobStatus = "queued"
	StatusProcessing JobStatus = "processing"
	StatusDone       JobStatus = "done"
	StatusFailed     JobStatus = "failed"
)

// Job interface for jobs
type Job interface {
	// Execute run job
	Execute() error

	// GetID returns ID
	GetID() string

	// GetStatus returns Status
	GetStatus() JobStatus
}

// BaseJob includes common fields
type BaseJob struct {
	ID     string
	Status JobStatus
}

// GetID returns ID
func (j BaseJob) GetID() string {
	return j.ID
}

// GetStatus returns Status
func (j BaseJob) GetStatus() JobStatus {
	return j.Status
}

// JobQueuer interface for Job Queues
type JobQueuer interface {
	AddJob(job Job)
	GetJobsStatuses() map[string]JobStatus
}

// JobQueue stores a queue of jobs and their statuses
type JobQueue struct {
	mu    sync.Mutex
	jobs  map[string]JobStatus
	queue chan Job
}

func NewJobQueue() *JobQueue {
	return &JobQueue{
		jobs:  make(map[string]JobStatus),
		queue: make(chan Job, 100), // buffer size is 100
	}
}

func (jq *JobQueue) AddJob(job Job) {
	jq.mu.Lock()
	jq.jobs[job.GetID()] = StatusQueued
	jq.mu.Unlock()
	jq.queue <- job
}

// UpdateStatus updates job status
func (jq *JobQueue) UpdateStatus(jobID string, status JobStatus) {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	jq.jobs[jobID] = status
}

// GetJobsStatuses returns jobs statuses
func (jq *JobQueue) GetJobsStatuses() map[string]JobStatus {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	list := make(map[string]JobStatus, len(jq.jobs))
	for id, status := range jq.jobs {
		list[id] = status
	}
	return list
}

// Worker processes tasks from the queue
func Worker(id int, jq *JobQueue) {
	for job := range jq.queue {
		jobID := job.GetID()

		log.Printf("Worker %d: job %s is processing", id, jobID)
		jq.UpdateStatus(jobID, StatusProcessing)

		err := job.Execute()
		if err != nil {
			log.Printf("Worker %d: failed job %s: %v", id, jobID, err)
			jq.UpdateStatus(jobID, StatusFailed)
		} else {
			jq.UpdateStatus(jobID, StatusDone)
			log.Printf("Worker %d: successful job %s", id, jobID)
		}
	}
}
