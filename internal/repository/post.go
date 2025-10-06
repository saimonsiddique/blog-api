package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saimonsiddique/blog-api/internal/domain"
)

type PostRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
	return &PostRepository{db: db}
}

// Create creates a new post
func (r *PostRepository) Create(ctx context.Context, post *domain.Post) error {
	query := `
		INSERT INTO posts (author_id, title, slug, content, excerpt, status, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, uuid, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		post.AuthorID,
		post.Title,
		post.Slug,
		post.Content,
		post.Excerpt,
		post.Status,
		post.PublishedAt,
	).Scan(&post.ID, &post.UUID, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		if err.Error() == `ERROR: duplicate key value violates unique constraint "posts_slug_key" (SQLSTATE 23505)` {
			return domain.ErrSlugTaken
		}
		return err
	}

	return nil
}

// GetByUUID retrieves a post by UUID with author information
func (r *PostRepository) GetByUUID(ctx context.Context, postUUID uuid.UUID) (*domain.PostWithAuthor, error) {
	query := `
		SELECT
			p.id, p.uuid, p.author_id, p.title, p.slug, p.content, p.excerpt,
			p.status, p.published_at, p.created_at, p.updated_at,
			u.uuid, u.username
		FROM posts p
		INNER JOIN users u ON p.author_id = u.id
		WHERE p.uuid = $1
	`

	var post domain.PostWithAuthor
	err := r.db.QueryRow(ctx, query, postUUID).Scan(
		&post.ID,
		&post.UUID,
		&post.AuthorID,
		&post.Title,
		&post.Slug,
		&post.Content,
		&post.Excerpt,
		&post.Status,
		&post.PublishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Author.UUID,
		&post.Author.Username,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}
		return nil, err
	}

	return &post, nil
}

// GetBySlug retrieves a post by slug with author information
func (r *PostRepository) GetBySlug(ctx context.Context, slug string) (*domain.PostWithAuthor, error) {
	query := `
		SELECT
			p.id, p.uuid, p.author_id, p.title, p.slug, p.content, p.excerpt,
			p.status, p.published_at, p.created_at, p.updated_at,
			u.uuid, u.username
		FROM posts p
		INNER JOIN users u ON p.author_id = u.id
		WHERE p.slug = $1
	`

	var post domain.PostWithAuthor
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&post.ID,
		&post.UUID,
		&post.AuthorID,
		&post.Title,
		&post.Slug,
		&post.Content,
		&post.Excerpt,
		&post.Status,
		&post.PublishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Author.UUID,
		&post.Author.Username,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}
		return nil, err
	}

	return &post, nil
}

// List retrieves posts with filters and pagination
func (r *PostRepository) List(ctx context.Context, req domain.ListPostsRequest) ([]domain.PostWithAuthor, int, error) {
	// Build query with filters
	query := `
		SELECT
			p.id, p.uuid, p.author_id, p.title, p.slug, p.content, p.excerpt,
			p.status, p.published_at, p.created_at, p.updated_at,
			u.uuid, u.username
		FROM posts p
		INNER JOIN users u ON p.author_id = u.id
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM posts p INNER JOIN users u ON p.author_id = u.id WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	// Add filters
	if req.Status != nil {
		query += ` AND p.status = $` + string(rune(argIndex+'0'))
		countQuery += ` AND p.status = $` + string(rune(argIndex+'0'))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.AuthorID != nil {
		// Get user ID from UUID
		var authorID int
		err := r.db.QueryRow(ctx, `SELECT id FROM users WHERE uuid = $1`, *req.AuthorID).Scan(&authorID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return []domain.PostWithAuthor{}, 0, nil
			}
			return nil, 0, err
		}

		query += ` AND p.author_id = $` + string(rune(argIndex+'0'))
		countQuery += ` AND p.author_id = $` + string(rune(argIndex+'0'))
		args = append(args, authorID)
		argIndex++
	}

	// Get total count
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Add ordering and pagination
	query += ` ORDER BY p.created_at DESC`

	if req.Limit > 0 {
		query += ` LIMIT $` + string(rune(argIndex+'0'))
		args = append(args, req.Limit)
		argIndex++
	}

	if req.Page > 1 && req.Limit > 0 {
		offset := (req.Page - 1) * req.Limit
		query += ` OFFSET $` + string(rune(argIndex+'0'))
		args = append(args, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []domain.PostWithAuthor
	for rows.Next() {
		var post domain.PostWithAuthor
		err := rows.Scan(
			&post.ID,
			&post.UUID,
			&post.AuthorID,
			&post.Title,
			&post.Slug,
			&post.Content,
			&post.Excerpt,
			&post.Status,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Author.UUID,
			&post.Author.Username,
		)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if posts == nil {
		posts = []domain.PostWithAuthor{}
	}

	return posts, totalCount, nil
}

// Update updates a post
func (r *PostRepository) Update(ctx context.Context, postUUID uuid.UUID, updates map[string]interface{}) (*domain.Post, error) {
	// Build dynamic update query
	query := `UPDATE posts SET `
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		if argIndex > 1 {
			query += `, `
		}
		query += field + ` = $` + string(rune(argIndex+'0'))
		args = append(args, value)
		argIndex++
	}

	query += `, updated_at = CURRENT_TIMESTAMP WHERE uuid = $` + string(rune(argIndex+'0'))
	args = append(args, postUUID)
	query += ` RETURNING id, uuid, author_id, title, slug, content, excerpt, status, published_at, created_at, updated_at`

	var post domain.Post
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&post.ID,
		&post.UUID,
		&post.AuthorID,
		&post.Title,
		&post.Slug,
		&post.Content,
		&post.Excerpt,
		&post.Status,
		&post.PublishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}
		if err.Error() == `ERROR: duplicate key value violates unique constraint "posts_slug_key" (SQLSTATE 23505)` {
			return nil, domain.ErrSlugTaken
		}
		return nil, err
	}

	return &post, nil
}

// Delete deletes a post
func (r *PostRepository) Delete(ctx context.Context, postUUID uuid.UUID) error {
	query := `DELETE FROM posts WHERE uuid = $1`

	result, err := r.db.Exec(ctx, query, postUUID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrPostNotFound
	}

	return nil
}

// IsAuthor checks if a user is the author of a post
func (r *PostRepository) IsAuthor(ctx context.Context, postUUID uuid.UUID, userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE uuid = $1 AND author_id = $2)`

	var exists bool
	err := r.db.QueryRow(ctx, query, postUUID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
