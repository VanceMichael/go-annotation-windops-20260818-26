package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"windops/internal/domain/outbox"
	"windops/internal/store"
)

type Publisher interface {
	Publish(context.Context, string, []byte) error
}
type LogPublisher struct{}

func (LogPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if topic == "" || len(payload) == 0 {
		return fmt.Errorf("topic and payload are required")
	}
	return nil
}

type OutboxWorker struct {
	db        *store.Database
	publisher Publisher
	interval  time.Duration
	batch     int
	wg        sync.WaitGroup
}

func NewOutboxWorker(db *store.Database, publisher Publisher, interval time.Duration, batch int) *OutboxWorker {
	return &OutboxWorker{db: db, publisher: publisher, interval: interval, batch: batch}
}
func (w *OutboxWorker) Run(ctx context.Context) {
	w.wg.Add(1)
	defer w.wg.Done()
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = w.Process(ctx)
		}
	}
}
func (w *OutboxWorker) Wait() { w.wg.Wait() }
func (w *OutboxWorker) Process(ctx context.Context) error {
	rows, err := w.db.SQL().QueryContext(ctx, `SELECT data_json FROM outbox_jobs WHERE status IN ('pending','retry') ORDER BY updated_at,id LIMIT ?`, w.batch)
	if err != nil {
		return err
	}
	defer rows.Close()
	jobs := []outbox.Job{}
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return err
		}
		var job outbox.Job
		if err := json.Unmarshal(raw, &job); err != nil {
			return err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, job := range jobs {
		if err := w.processOne(ctx, job); err != nil && !errors.Is(err, context.Canceled) {
			continue
		}
	}
	return ctx.Err()
}
func (w *OutboxWorker) processOne(ctx context.Context, job outbox.Job) error {
	if time.Now().UTC().Before(job.AvailableAt) {
		return nil
	}
	err := w.publisher.Publish(ctx, job.Topic, []byte(job.Payload))
	previous := job.Version
	job.Version++
	job.UpdatedAt = time.Now().UTC()
	if err == nil {
		job.Status = outbox.StatusSucceeded
		job.LastError = ""
	} else {
		job.Attempts++
		job.LastError = err.Error()
		if job.Attempts >= 5 {
			job.Status = outbox.StatusPermanentFailure
		} else {
			job.Status = outbox.StatusRetry
			job.AvailableAt = job.UpdatedAt.Add(time.Duration(1<<min(job.Attempts, 8)) * time.Second)
		}
	}
	return store.NewJobRepository(w.db).Update(ctx, job, previous)
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
