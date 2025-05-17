package job

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockJob implements the Job interface using testify/mock
type MockJob struct {
	mock.Mock
	BaseJob
}

// Execute mocks the job execution logic
func (m *MockJob) Execute() error {
	args := m.Called()
	return args.Error(0)
}

// --- Tests ---

func TestJobQueue_AddJobAndGetStatuses(t *testing.T) {
	jq := NewJobQueue()

	job := &MockJob{}
	job.ID = "job-1"
	job.Status = StatusQueued

	// Expect Execute to never be called in this test
	job.On("Execute").Return(nil)

	jq.AddJob(job)

	// Wait a moment to ensure AddJob sends to queue
	time.Sleep(10 * time.Millisecond)

	statuses := jq.GetJobsStatuses()
	assert.Equal(t, 1, len(statuses))
	assert.Equal(t, StatusQueued, statuses["job-1"])
}

func TestJobQueue_UpdateStatus(t *testing.T) {
	jq := NewJobQueue()

	jq.UpdateStatus("job-2", StatusProcessing)
	statuses := jq.GetJobsStatuses()

	assert.Equal(t, StatusProcessing, statuses["job-2"])
}

func TestWorker_Success(t *testing.T) {
	jq := NewJobQueue()

	job := &MockJob{}
	job.ID = "job-3"

	// Simulate successful execution
	job.On("Execute").Return(nil)

	jq.AddJob(job)

	// Start worker in background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		Worker(1, jq)
	}()

	// Let worker process job
	time.Sleep(50 * time.Millisecond)

	// Stop the queue to exit worker
	close(jq.queue)
	wg.Wait()

	statuses := jq.GetJobsStatuses()
	assert.Equal(t, StatusDone, statuses["job-3"])
	job.AssertExpectations(t)
}

func TestWorker_Failure(t *testing.T) {
	jq := NewJobQueue()

	job := &MockJob{}
	job.ID = "job-4"

	// Simulate failed execution
	job.On("Execute").Return(errors.New("error occurred"))

	jq.AddJob(job)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		Worker(2, jq)
	}()

	time.Sleep(50 * time.Millisecond)
	close(jq.queue)
	wg.Wait()

	statuses := jq.GetJobsStatuses()
	assert.Equal(t, StatusFailed, statuses["job-4"])
	job.AssertExpectations(t)
}
