package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/saimonsiddique/blog-api/internal/domain"
)

type PostPublisher struct {
	queue *RabbitMQ
}

func NewPostPublisher(queue *RabbitMQ) *PostPublisher {
	return &PostPublisher{
		queue: queue,
	}
}

func (p *PostPublisher) PublishPostPublishEvent(ctx context.Context, event *domain.PostPublishEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.queue.Publish(ctx, domain.QueuePostPublish, body)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}
