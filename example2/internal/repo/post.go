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
	ListPosts(ctx context.Context, limit, offset int) ([]model.Post, error)
	CountPosts(ctx context.Context) (int64, error)
	GetPost(ctx context.Context, postID uint64) (*model.Post, error)
	GetPostsByUserID(ctx context.Context, userID uint64) ([]model.Post, error)
	GetPostsByUserIDs(ctx context.Context, userIDs []uint64) (map[uint64][]model.Post, error)
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

// ListPosts returns a paginated list of posts ordered by post_id descending.
func (r *PostRepo) ListPosts(ctx context.Context, limit, offset int) ([]model.Post, error) {
	var posts []model.Post
	err := r.db.WithContext(ctx).
		Order("post_id desc").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// CountPosts returns the total number of posts in the database.
func (r *PostRepo) CountPosts(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.Post{}).Count(&total).Error
	return total, err
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

// GetPostsByUserIDs returns posts for the given userIDs grouped by user ID,
// loaded in a single query to avoid per-user (N+1) lookups.
func (r *PostRepo) GetPostsByUserIDs(ctx context.Context, userIDs []uint64) (map[uint64][]model.Post, error) {
	grouped := make(map[uint64][]model.Post)
	if len(userIDs) == 0 {
		return grouped, nil
	}

	var posts []model.Post
	if err := r.db.WithContext(ctx).Where("id_user IN ?", userIDs).Find(&posts).Error; err != nil {
		return nil, err
	}

	for _, post := range posts {
		grouped[post.IDUser] = append(grouped[post.IDUser], post)
	}
	return grouped, nil
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
