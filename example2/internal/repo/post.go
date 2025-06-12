package repo

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/pilinux/gorest/example2/internal/database/model"
)

// PostRepo provides methods for post-related database operations.
type PostRepo struct {
	db *gorm.DB
}

// NewPostRepo returns a new PostRepo with the given database connection.
func NewPostRepo(conn *gorm.DB) *PostRepo {
	return &PostRepo{
		db: conn,
	}
}

// PostRepository defines the contract for post-related operations.
type PostRepository interface {
	GetPosts(ctx context.Context) ([]model.Post, error)
	GetPost(ctx context.Context, postID uint64) (*model.Post, error)
	GetPostsByUserID(ctx context.Context, userID uint64) ([]model.Post, error)
	CreatePost(ctx context.Context, post *model.Post) error
	UpdatePost(ctx context.Context, post *model.Post) error
	DeletePost(ctx context.Context, postID uint64) error
	DeletePostsByUserID(ctx context.Context, userID uint64) error
	DeletePostsByAuthID(ctx context.Context, authID uint64) error
}

// Compile-time check:
var _ PostRepository = (*PostRepo)(nil)

// GetPosts returns all posts from the database.
func (r *PostRepo) GetPosts(ctx context.Context) ([]model.Post, error) {
	var posts []model.Post
	if err := r.db.WithContext(ctx).Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// GetPost returns a post with the given postID from the database.
func (r *PostRepo) GetPost(ctx context.Context, postID uint64) (*model.Post, error) {
	var post model.Post
	if postID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if err := r.db.WithContext(ctx).Where("post_id = ?", postID).First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

// GetPostsByUserID returns all posts for a given userID from the database.
func (r *PostRepo) GetPostsByUserID(ctx context.Context, userID uint64) ([]model.Post, error) {
	var posts []model.Post
	if userID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if err := r.db.WithContext(ctx).Where("id_user = ?", userID).Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// CreatePost creates a new post in the database.
func (r *PostRepo) CreatePost(ctx context.Context, post *model.Post) error {
	tNow := time.Now()
	post.PostID = 0 // auto-increment
	post.CreatedAt = tNow.Unix()
	post.UpdatedAt = tNow.Unix()
	return r.db.WithContext(ctx).Create(post).Error
}

// UpdatePost updates an existing post in the database.
func (r *PostRepo) UpdatePost(ctx context.Context, post *model.Post) error {
	post.UpdatedAt = time.Now().Unix()
	return r.db.WithContext(ctx).Save(post).Error
}

// DeletePost deletes a post with the given postID from the database.
func (r *PostRepo) DeletePost(ctx context.Context, postID uint64) error {
	return r.db.WithContext(ctx).Where("post_id = ?", postID).Delete(&model.Post{}).Error
}

// DeletePostsByUserID deletes all posts for a given userID from the database.
func (r *PostRepo) DeletePostsByUserID(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Where("id_user = ?", userID).Delete(&model.Post{}).Error
}

// DeletePostsByAuthID deletes all posts for a given authID from the database.
func (r *PostRepo) DeletePostsByAuthID(ctx context.Context, authID uint64) error {
	return r.db.WithContext(ctx).Where("id_auth = ?", authID).Delete(&model.Post{}).Error
}
