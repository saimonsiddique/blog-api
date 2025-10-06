package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/saimonsiddique/blog-api/internal/domain"
	"github.com/saimonsiddique/blog-api/internal/pkg/slug"
	"github.com/saimonsiddique/blog-api/internal/queue"
	"github.com/saimonsiddique/blog-api/internal/repository"
)

type PostService struct {
	postRepo      *repository.PostRepository
	userRepo      *repository.UserRepository
	postPublisher *queue.PostPublisher
}

func NewPostService(postRepo *repository.PostRepository, userRepo *repository.UserRepository, postPublisher *queue.PostPublisher) *PostService {
	return &PostService{
		postRepo:      postRepo,
		userRepo:      userRepo,
		postPublisher: postPublisher,
	}
}

// Create creates a new post
func (s *PostService) Create(ctx context.Context, userUUID uuid.UUID, req domain.CreatePostRequest) (*domain.PostResponse, error) {
	// Get user by UUID
	user, err := s.userRepo.GetByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	// Generate slug from title
	postSlug := slug.Generate(req.Title)

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = domain.PostStatusDraft
	}

	// Set published_at if status is published
	var publishedAt *time.Time
	if status == domain.PostStatusPublished {
		now := time.Now()
		publishedAt = &now
	}

	// Create post
	post := &domain.Post{
		AuthorID:    user.ID,
		Title:       req.Title,
		Slug:        postSlug,
		Content:     req.Content,
		Excerpt:     req.Excerpt,
		Status:      status,
		PublishedAt: publishedAt,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}

	// Return response
	return &domain.PostResponse{
		UUID:        post.UUID,
		Title:       post.Title,
		Slug:        post.Slug,
		Content:     post.Content,
		Excerpt:     post.Excerpt,
		Status:      post.Status,
		PublishedAt: post.PublishedAt,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Author: domain.PostAuthor{
			UUID:     user.UUID,
			Username: user.Username,
		},
	}, nil
}

// GetByUUID retrieves a post by UUID
func (s *PostService) GetByUUID(ctx context.Context, postUUID uuid.UUID) (*domain.PostResponse, error) {
	post, err := s.postRepo.GetByUUID(ctx, postUUID)
	if err != nil {
		return nil, err
	}

	return &domain.PostResponse{
		UUID:        post.UUID,
		Title:       post.Title,
		Slug:        post.Slug,
		Content:     post.Content,
		Excerpt:     post.Excerpt,
		Status:      post.Status,
		PublishedAt: post.PublishedAt,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Author:      post.Author,
	}, nil
}

// GetBySlug retrieves a post by slug
func (s *PostService) GetBySlug(ctx context.Context, slug string) (*domain.PostResponse, error) {
	post, err := s.postRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return &domain.PostResponse{
		UUID:        post.UUID,
		Title:       post.Title,
		Slug:        post.Slug,
		Content:     post.Content,
		Excerpt:     post.Excerpt,
		Status:      post.Status,
		PublishedAt: post.PublishedAt,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Author:      post.Author,
	}, nil
}

// List retrieves posts with filters and pagination
func (s *PostService) List(ctx context.Context, req domain.ListPostsRequest) (*domain.ListPostsResponse, error) {
	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	posts, totalCount, err := s.postRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	postResponses := make([]domain.PostResponse, len(posts))
	for i, post := range posts {
		postResponses[i] = domain.PostResponse{
			UUID:        post.UUID,
			Title:       post.Title,
			Slug:        post.Slug,
			Content:     post.Content,
			Excerpt:     post.Excerpt,
			Status:      post.Status,
			PublishedAt: post.PublishedAt,
			CreatedAt:   post.CreatedAt,
			UpdatedAt:   post.UpdatedAt,
			Author:      post.Author,
		}
	}

	return &domain.ListPostsResponse{
		Posts:      postResponses,
		TotalCount: totalCount,
		Page:       req.Page,
		Limit:      req.Limit,
	}, nil
}

// Update updates a post
func (s *PostService) Update(ctx context.Context, userUUID uuid.UUID, postUUID uuid.UUID, req domain.UpdatePostRequest) (*domain.PostResponse, error) {
	// Get user by UUID
	user, err := s.userRepo.GetByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	// Check if user is the author
	isAuthor, err := s.postRepo.IsAuthor(ctx, postUUID, user.ID)
	if err != nil {
		return nil, err
	}
	if !isAuthor {
		return nil, domain.ErrForbidden
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
		updates["slug"] = slug.Generate(*req.Title)
	}

	if req.Content != nil {
		updates["content"] = *req.Content
	}

	if req.Excerpt != nil {
		updates["excerpt"] = *req.Excerpt
	}

	if req.Status != nil {
		// Get current post to check status transitions
		currentPost, err := s.postRepo.GetByUUID(ctx, postUUID)
		if err != nil {
			return nil, err
		}

		// Handle publish status change via queue
		if *req.Status == domain.PostStatusPublished {
			// Check if already published
			if currentPost.Status == domain.PostStatusPublished {
				return nil, domain.ErrPostAlreadyPublished
			}

			// Enqueue publish event
			event := &domain.PostPublishEvent{
				PostUUID:     postUUID.String(),
				AuthorUUID:   userUUID.String(),
				RequestedAt:  time.Now(),
				ScheduledFor: req.ScheduledFor,
			}

			if err := s.postPublisher.PublishPostPublishEvent(ctx, event); err != nil {
				return nil, err
			}

			// Don't update status directly - worker will handle it
			// Return current post state
			post, err := s.postRepo.GetByUUID(ctx, postUUID)
			if err != nil {
				return nil, err
			}

			return &domain.PostResponse{
				UUID:        post.UUID,
				Title:       post.Title,
				Slug:        post.Slug,
				Content:     post.Content,
				Excerpt:     post.Excerpt,
				Status:      post.Status,
				PublishedAt: post.PublishedAt,
				CreatedAt:   post.CreatedAt,
				UpdatedAt:   post.UpdatedAt,
				Author:      post.Author,
			}, nil
		} else {
			// Validate status transitions
			if err := s.validateStatusChange(currentPost.Status, *req.Status); err != nil {
				return nil, err
			}

			updates["status"] = *req.Status

			// Clear published_at when changing to draft or archived
			if *req.Status == domain.PostStatusDraft || *req.Status == domain.PostStatusArchived {
				updates["published_at"] = nil
			}
		}
	}

	// Update post
	updatedPost, err := s.postRepo.Update(ctx, postUUID, updates)
	if err != nil {
		return nil, err
	}

	// Get full post with author info
	post, err := s.postRepo.GetByUUID(ctx, postUUID)
	if err != nil {
		return nil, err
	}

	return &domain.PostResponse{
		UUID:        post.UUID,
		Title:       post.Title,
		Slug:        post.Slug,
		Content:     post.Content,
		Excerpt:     post.Excerpt,
		Status:      post.Status,
		PublishedAt: post.PublishedAt,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   updatedPost.UpdatedAt,
		Author:      post.Author,
	}, nil
}

// validateStatusChange validates if a status transition is allowed
func (s *PostService) validateStatusChange(currentStatus, newStatus domain.PostStatus) error {
	// Allow transitions to the same status (no-op)
	if currentStatus == newStatus {
		return nil
	}

	// Define allowed transitions
	allowedTransitions := map[domain.PostStatus][]domain.PostStatus{
		domain.PostStatusDraft:     {domain.PostStatusPublished, domain.PostStatusArchived},
		domain.PostStatusPublished: {domain.PostStatusDraft, domain.PostStatusArchived},
		domain.PostStatusArchived:  {domain.PostStatusDraft},
	}

	allowed, exists := allowedTransitions[currentStatus]
	if !exists {
		return domain.ErrInvalidStatusChange
	}

	for _, allowedStatus := range allowed {
		if allowedStatus == newStatus {
			return nil
		}
	}

	return domain.ErrInvalidStatusChange
}

// Delete deletes a post
func (s *PostService) Delete(ctx context.Context, userUUID uuid.UUID, postUUID uuid.UUID) error {
	// Get user by UUID
	user, err := s.userRepo.GetByUUID(ctx, userUUID)
	if err != nil {
		return err
	}

	// Check if user is the author
	isAuthor, err := s.postRepo.IsAuthor(ctx, postUUID, user.ID)
	if err != nil {
		return err
	}
	if !isAuthor {
		return domain.ErrForbidden
	}

	return s.postRepo.Delete(ctx, postUUID)
}
