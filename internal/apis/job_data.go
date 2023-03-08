package apis

import (
	"context"
	"csc-code-test/internal/logger"
	"csc-code-test/internal/settings"
	"sync"
	"time"
)

const (
	JobStatusPending   = "pending"
	JobStatusProcessed = "processed"
)

type JobModel struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Data   []int  `json:"data"`
	Status string `json:"status"`
	Result int    `json:"result"`
}

type JobQueueManager struct {
	itemsPendingJobs   []JobModel
	itemsProcessedJobs []JobModel
	itemsJobMutex      sync.RWMutex

	waitJobWorker   sync.WaitGroup
	workerCtx       context.Context
	workerCancelCtx context.CancelFunc
	flagStop        bool
}

var initializeOnce sync.Once

var JobQueueManagerShared *JobQueueManager

func InitializeJobQueueManager() {
	initializeOnce.Do(func() {
		JobQueueManagerShared = new(JobQueueManager)
		JobQueueManagerShared.Initialize()
	})
}

func (job *JobQueueManager) Initialize() {
	job.itemsPendingJobs = make([]JobModel, 0)
	job.itemsProcessedJobs = make([]JobModel, 0)
}

func (job *JobQueueManager) StartWorker() {
	job.workerCtx, job.workerCancelCtx = context.WithCancel(context.Background())

	job.waitJobWorker.Add(1)
	go job.startBackgroundWorker()

	logger.Logger.Info("job worker started")
}

func (job *JobQueueManager) StopWorker() {
	logger.Logger.Info("stopping job worker...")

	job.flagStop = true
	job.workerCancelCtx()
	job.waitJobWorker.Wait()

	logger.Logger.Info("job worker stopped")
}

func (job *JobQueueManager) startBackgroundWorker() {
	wait := time.Duration(settings.Settings.SecondsToWait) * time.Second

	for {
		select {
		case <-job.workerCtx.Done():
			break
		case <-time.After(wait):
			job.processJobQueue()
			break
		}

		if job.flagStop {
			break
		}
	}

	job.waitJobWorker.Done()
}

func (job *JobQueueManager) processJobQueue() {
	job.itemsJobMutex.Lock()

	if len(job.itemsPendingJobs) > 0 {

		// FIFO
		currentJob := job.itemsPendingJobs[0]

		job.itemsPendingJobs = job.itemsPendingJobs[1:]

		// 'Processing a job is equal to taking the values of the "data" key in the user input
		// and adding all values within it to each other' :  the sum of all values?

		sumData := 0
		for _, value := range currentJob.Data {
			sumData = sumData + value
		}

		currentJob.Result = sumData
		currentJob.Status = JobStatusProcessed

		job.itemsProcessedJobs = append(job.itemsProcessedJobs, currentJob)
	}

	job.itemsJobMutex.Unlock()
}

func (job *JobQueueManager) PushJobTask(name string, data []int) {
	job.itemsJobMutex.Lock()

	// 'ID will be the parameter representing the priority' : this ID represents the push order to be processed ?
	newId := len(job.itemsPendingJobs) + 1

	jobModel := JobModel{
		Id:     newId,
		Name:   name,
		Data:   data,
		Status: JobStatusPending,
	}

	// FIFO
	job.itemsPendingJobs = append(job.itemsPendingJobs, jobModel)

	job.itemsJobMutex.Unlock()
}

func (job *JobQueueManager) ListJobs(status string) []JobModel {
	if status == JobStatusPending {
		return job.itemsPendingJobs
	}

	if status == JobStatusProcessed {
		return job.itemsProcessedJobs
	}

	job.itemsJobMutex.RLock()
	allJobs := append(job.itemsProcessedJobs, job.itemsPendingJobs...)
	job.itemsJobMutex.RUnlock()

	return allJobs
}
