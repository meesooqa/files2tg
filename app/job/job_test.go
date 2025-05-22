package job

import (
	"errors"
	"fmt"
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

// TestClearOnEmptyQueue проверяет, что Clear на пустой очереди не паникует и сбрасывает map
func TestClearOnEmptyQueue(t *testing.T) {
	jq := NewJobQueue()
	if len(jq.GetJobsStatuses()) != 0 {
		t.Fatalf("ожидалось 0 задач, получили %d", len(jq.GetJobsStatuses()))
	}

	jq.Clear()

	if len(jq.GetJobsStatuses()) != 0 {
		t.Errorf("после Clear ожидается 0 задач, получили %d", len(jq.GetJobsStatuses()))
	}
}

// TestClearDrainsPendingJobs проверяет, что Clear очищает и канал, и map
func TestClearDrainsPendingJobs(t *testing.T) {
	jq := NewJobQueue()
	for i := 0; i < 5; i++ {
		job := &MockJob{}
		job.ID = fmt.Sprintf("%s-%d", "cjob", i)
		jq.AddJob(job)
	}
	if len(jq.GetJobsStatuses()) != 5 {
		t.Fatalf("ожидалось 5 задач до Clear, получили %d", len(jq.GetJobsStatuses()))
	}

	jq.Clear()

	if len(jq.GetJobsStatuses()) != 0 {
		t.Errorf("после Clear() map не пустой, осталось %d", len(jq.GetJobsStatuses()))
	}

	done := make(chan struct{})
	go func() {
		// поскольку канал дренирован, блокировки быть не должно
		job := &MockJob{}
		job.ID = "X"
		jq.queue <- job
		done <- struct{}{}
	}()

	select {
	case <-done:
		// ok
	case <-time.After(100 * time.Millisecond):
		t.Error("поток заблокирован при отправке в канал после Clear()")
	}
}

// TestClearDuringProcessing проверяет, что Clear безопасен при одновременном работе Worker
func TestClearDuringProcessing(t *testing.T) {
	jq := NewJobQueue()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		Worker(1, jq)
	}()

	// добавляем задачи, которые завершаются с ошибкой и без
	job1 := &MockJob{}
	job1.ID = "ok"
	job1.On("Execute").Return(nil)
	jq.AddJob(job1)
	job2 := &MockJob{}
	job2.ID = "fail"
	job2.On("Execute").Return(nil)
	jq.AddJob(job2)

	// дадим воркеру немного поработать
	time.Sleep(5 * time.Millisecond)

	// теперь чистим
	jq.Clear()

	// закрываем очередь, чтобы воркер вышел
	close(jq.queue) // если поле queue не экспортируется, замените на закрытие через метод

	wg.Wait()

	// после завершения всё равно должно быть пусто
	if len(jq.GetJobsStatuses()) != 0 {
		t.Errorf("после Clear + обработка задачи остались в map: %v", jq.GetJobsStatuses())
	}
}
