package services

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrJobQueueIsFull = errors.New("job queue is full")
)

type Job func(ctx context.Context)

type JobQueueService struct {
	jobs   chan Job
	resume chan struct{}
	paused int32
	wg     sync.WaitGroup
}

func NewJobQueueService(ctx context.Context, capacity, workers int) *JobQueueService {
	service := &JobQueueService{
		jobs:   make(chan Job, capacity),
		resume: make(chan struct{}),
		wg:     sync.WaitGroup{},
	}
	service.start(ctx, workers)

	return service
}

func (jqs *JobQueueService) start(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ {
		jqs.wg.Add(1)

		go func() {
			defer jqs.wg.Done()

			for {
				select {
				case job, ok := <-jqs.jobs:
					if !ok {
						return
					}

					if atomic.LoadInt32(&jqs.paused) == 1 {
						<-jqs.resume
					}

					job(ctx)
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func (jqs *JobQueueService) Enqueue(job Job) {
	jqs.jobs <- job
	// todo think about handling overflow capacity
	//select {
	//case jqs.jobs <- job:
	//	return nil
	//default:
	//	return ErrJobQueueIsFull
	//}
}

func (jqs *JobQueueService) ScheduleJob(job Job, delay time.Duration) {
	time.AfterFunc(delay, func() {
		jqs.jobs <- job
	})
}

func (jqs *JobQueueService) Pause() {
	atomic.StoreInt32(&jqs.paused, 1)
}

func (jqs *JobQueueService) Resume() {
	if atomic.CompareAndSwapInt32(&jqs.paused, 1, 0) {
		close(jqs.resume)
		jqs.resume = make(chan struct{})
	}
}

func (jqs *JobQueueService) PauseAndResume(delay time.Duration) {
	jqs.Pause()
	time.AfterFunc(delay, func() {
		jqs.Resume()
	})
}

func (jqs *JobQueueService) Shutdown() {
	close(jqs.jobs)
	jqs.wg.Wait()
}