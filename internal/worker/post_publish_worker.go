package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/saimonsiddique/blog-api/internal/domain"
	"github.com/saimonsiddique/blog-api/internal/queue"
	"github.com/sirupsen/logrus"
)

type PostPublishWorker struct {
	queue  *queue.RabbitMQ
	db     *pgxpool.Pool
	logger *logrus.Logger
}

func NewPostPublishWorker(queue *queue.RabbitMQ, db *pgxpool.Pool, logger *logrus.Logger) *PostPublishWorker {
	return &PostPublishWorker{
		queue:  queue,
		db:     db,
		logger: logger,
	}
}

func (w *PostPublishWorker) Start(ctx context.Context) error {
	// Declare queue
	err := w.queue.DeclareQueue(domain.QueuePostPublish)
	if err != nil {
		return err
	}

	// Start consuming
	msgs, err := w.queue.Consume(domain.QueuePostPublish)
	if err != nil {
		return err
	}

	w.logger.Info("Post publish worker started")

	go func() {
		for {
			select {
			case <-ctx.Done():
				w.logger.Info("Post publish worker stopped")
				return
			case msg := <-msgs:
				w.processMessage(msg)
			}
		}
	}()

	return nil
}

func (w *PostPublishWorker) processMessage(msg amqp.Delivery) {
	var event domain.PostPublishEvent
	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		w.logger.Errorf("Failed to unmarshal message: %v", err)
		msg.Nack(false, false) // Don't requeue invalid messages
		return
	}

	w.logger.Infof("Processing post publish event for post: %s", event.PostUUID)

	// Check if scheduled for future
	if event.ScheduledFor != nil && event.ScheduledFor.After(time.Now()) {
		delay := time.Until(*event.ScheduledFor)
		w.logger.Infof("Post %s scheduled for %v, waiting %v", event.PostUUID, event.ScheduledFor, delay)
		time.Sleep(delay)
	}

	// Publish the post
	err = w.publishPost(context.Background(), event.PostUUID)
	if err != nil {
		w.logger.Errorf("Failed to publish post %s: %v", event.PostUUID, err)
		msg.Nack(false, true) // Requeue on failure
		return
	}

	w.logger.Infof("Successfully published post: %s", event.PostUUID)
	msg.Ack(false)
}

func (w *PostPublishWorker) publishPost(ctx context.Context, postUUID string) error {
	query := `
		UPDATE posts
		SET status = 'published',
		    published_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE uuid = $1 AND status = 'draft'
	`

	result, err := w.db.Exec(ctx, query, postUUID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		w.logger.Warnf("Post %s not found or already published", postUUID)
	}

	return nil
}
