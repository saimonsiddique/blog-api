package domain

import "time"

// PostPublishEvent represents a post publish event to be queued
type PostPublishEvent struct {
	PostUUID      string    `json:"postUuid"`
	AuthorUUID    string    `json:"authorUuid"`
	RequestedAt   time.Time `json:"requestedAt"`
	ScheduledFor  *time.Time `json:"scheduledFor,omitempty"`
}

// QueueName constants
const (
	QueuePostPublish = "post.publish"
)
