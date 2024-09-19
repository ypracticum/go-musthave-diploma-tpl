package services

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
	"fmt"
)

// Определение пользовательских ошибок.
var (
	ErrJobQueueIsFull = errors.New("очередь заданий заполнена")
	ErrJobQueueClosed = errors.New("очередь заданий закрыта")
)

// Job представляет собой функцию, выполняющуюся в очереди заданий.
type Job func(ctx context.Context)

// JobQueueService предоставляет функционал для управления очередью заданий.
type JobQueueService struct {
	jobs    chan Job        // Канал для очереди заданий.
	resume  chan struct{}   // Канал для возобновления выполнения заданий после паузы.
	paused  int32           // Флаг состояния паузы (1 - приостановлено, 0 - активно).
	wg      sync.WaitGroup  // Группа ожидания для отслеживания горутин.
	mu      sync.Mutex      // Мьютекс для защиты операций с каналом resume.
	closing int32           // Флаг закрытия очереди (1 - закрыта, 0 - активно).
}

// NewJobQueueService создает новый экземпляр JobQueueService.
// Параметры:
// - ctx: контекст для управления временем жизни сервиса.
// - capacity: емкость очереди заданий.
// - workers: количество воркеров, обрабатывающих задания.
func NewJobQueueService(ctx context.Context, capacity, workers int) *JobQueueService {
	service := &JobQueueService{
		jobs:   make(chan Job, capacity),
		resume: make(chan struct{}),
	}
	service.start(ctx, workers)

	return service
}

// start запускает заданное количество воркеров для обработки заданий.
func (jqs *JobQueueService) start(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ {
		jqs.wg.Add(1)

		go func(workerID int) {
			defer jqs.wg.Done()

			for {
				select {
				case job, ok := <-jqs.jobs:
					if !ok {
						// Канал закрыт, завершение воркера.
						return
					}

					// Проверка состояния паузы.
					if atomic.LoadInt32(&jqs.paused) == 1 {
						// Ожидание сигнала возобновления.
						<-jqs.resume
					}

					// Выполнение задания.
					job(ctx)
				case <-ctx.Done():
					// Завершение при отмене контекста.
					return
				}
			}
		}(i + 1)
	}
}

// Enqueue добавляет новое задание в очередь.
// Возвращает ошибку, если очередь заполнена или закрыта.
func (jqs *JobQueueService) Enqueue(job Job) error {
	// Проверка, закрыта ли очередь.
	if atomic.LoadInt32(&jqs.closing) == 1 {
		return ErrJobQueueClosed
	}

	select {
	case jqs.jobs <- job:
		return nil
	default:
		return ErrJobQueueIsFull
	}
}

// ScheduleJob планирует выполнение задания через заданную задержку.
func (jqs *JobQueueService) ScheduleJob(job Job, delay time.Duration) {
	time.AfterFunc(delay, func() {
		if err := jqs.Enqueue(job); err != nil {
			// Обработка ошибки добавления задания.
			// В случае использования этого кода в продакшене, можно заменить на вашу собственную обработку ошибок.
			fmt.Printf("Не удалось запланировать задание: %v\n", err)
		}
	})
}

// Pause приостанавливает выполнение заданий.
func (jqs *JobQueueService) Pause() {
	if atomic.CompareAndSwapInt32(&jqs.paused, 0, 1) {
		// Пауза была активирована.
	}
}

// Resume возобновляет выполнение заданий после паузы.
func (jqs *JobQueueService) Resume() {
	if atomic.CompareAndSwapInt32(&jqs.paused, 1, 0) {
		jqs.mu.Lock()
		defer jqs.mu.Unlock()
		// Закрытие текущего канала resume для освобождения блокированных воркеров.
		close(jqs.resume)
		// Создание нового канала resume для будущих пауз.
		jqs.resume = make(chan struct{})
	}
}

// PauseAndResume приостанавливает выполнение заданий на заданный промежуток времени, а затем возобновляет.
func (jqs *JobQueueService) PauseAndResume(delay time.Duration) {
	jqs.Pause()
	time.AfterFunc(delay, func() {
		jqs.Resume()
	})
}

// Shutdown корректно завершает работу очереди заданий.
// Закрывает канал заданий и ожидает завершения всех воркеров.
func (jqs *JobQueueService) Shutdown() {
	if atomic.CompareAndSwapInt32(&jqs.closing, 0, 1) {
		// Закрытие канала заданий.
		close(jqs.jobs)
		// Ожидание завершения всех воркеров.
		jqs.wg.Wait()
	}
}
