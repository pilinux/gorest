package repo

import (
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

// GetPosts returns all posts from the database.
func (r *PostRepo) GetPosts() ([]model.Post, error) {
	var posts []model.Post
	if err := r.db.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// GetPost returns a post with the given postID from the database.
func (r *PostRepo) GetPost(postID uint64) (*model.Post, error) {
	var post model.Post
	if postID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if err := r.db.Where("post_id = ?", postID).First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

// GetPostsByUserID returns all posts for a given userID from the database.
func (r *PostRepo) GetPostsByUserID(userID uint64) ([]model.Post, error) {
	var posts []model.Post
	if userID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if err := r.db.Where("id_user = ?", userID).Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// CreatePost creates a new post in the database.
func (r *PostRepo) CreatePost(post *model.Post) error {
	tNow := time.Now()
	post.PostID = 0 // auto-increment
	post.CreatedAt = tNow.Unix()
	post.UpdatedAt = tNow.Unix()
	return r.db.Create(post).Error
}

// UpdatePost updates an existing post in the database.
func (r *PostRepo) UpdatePost(post *model.Post) error {
	post.UpdatedAt = time.Now().Unix()
	return r.db.Save(post).Error
}

// DeletePost deletes a post with the given postID from the database.
func (r *PostRepo) DeletePost(postID uint64) error {
	return r.db.Where("post_id = ?", postID).Delete(&model.Post{}).Error
}

// DeletePostsByUserID deletes all posts for a given userID from the database.
func (r *PostRepo) DeletePostsByUserID(userID uint64) error {
	return r.db.Where("id_user = ?", userID).Delete(&model.Post{}).Error
}

// DeletePostsByAuthID deletes all posts for a given authID from the database.
func (r *PostRepo) DeletePostsByAuthID(authID uint64) error {
	return r.db.Where("id_auth = ?", authID).Delete(&model.Post{}).Error
}
