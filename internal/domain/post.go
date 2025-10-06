package domain

import (
	"time"

	"github.com/google/uuid"
)

// PostStatus represents the publication status of a post
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
	PostStatusArchived  PostStatus = "archived"
)

// Post represents a blog post
type Post struct {
	ID          int        `json:"id"`
	UUID        uuid.UUID  `json:"uuid"`
	AuthorID    int        `json:"authorId"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Content     string     `json:"content"`
	Excerpt     *string    `json:"excerpt,omitempty"`
	Status      PostStatus `json:"status"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// PostAuthor represents minimal author information for a post
type PostAuthor struct {
	UUID     uuid.UUID `json:"uuid"`
	Username string    `json:"username"`
}

// PostWithAuthor represents a post with author information
type PostWithAuthor struct {
	Post
	Author PostAuthor `json:"author"`
}

// CreatePostRequest represents the request to create a post
type CreatePostRequest struct {
	Title   string     `json:"title" validate:"required,min=3,max=255"`
	Content string     `json:"content" validate:"required,min=10"`
	Excerpt *string    `json:"excerpt" validate:"omitempty,max=500"`
	Status  PostStatus `json:"status" validate:"omitempty,oneof=draft published"`
}

// UpdatePostRequest represents the request to update a post
type UpdatePostRequest struct {
	Title        *string     `json:"title" validate:"omitempty,min=3,max=255"`
	Content      *string     `json:"content" validate:"omitempty,min=10"`
	Excerpt      *string     `json:"excerpt" validate:"omitempty,max=500"`
	Status       *PostStatus `json:"status" validate:"omitempty,oneof=draft published archived"`
	ScheduledFor *time.Time  `json:"scheduledFor" validate:"omitempty"`
}

// ListPostsRequest represents query parameters for listing posts
type ListPostsRequest struct {
	Status   *PostStatus `form:"status" validate:"omitempty,oneof=draft published archived"`
	AuthorID *uuid.UUID  `form:"authorId"`
	Page     int         `form:"page" validate:"omitempty,min=1"`
	Limit    int         `form:"limit" validate:"omitempty,min=1,max=100"`
}

// PostResponse represents a single post response
type PostResponse struct {
	UUID        uuid.UUID  `json:"uuid"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Content     string     `json:"content"`
	Excerpt     *string    `json:"excerpt,omitempty"`
	Status      PostStatus `json:"status"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Author      PostAuthor `json:"author"`
}

// ListPostsResponse represents the response for listing posts
type ListPostsResponse struct {
	Posts      []PostResponse `json:"posts"`
	TotalCount int            `json:"totalCount"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
}
