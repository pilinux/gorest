package service

import (
	"context"
	"errors"
	"net/http"

	gmodel "github.com/pilinux/gorest/database/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/pilinux/gorest/example2/internal/database/model"
	"github.com/pilinux/gorest/example2/internal/repo"
)

// PostService provides methods for post-related operations.
type PostService struct {
	postRepo repo.PostRepository
	userRepo repo.UserRepository
}

// NewPostService returns a new PostService instance.
func NewPostService(postRepo repo.PostRepository, userRepo repo.UserRepository) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// GetPosts retrieves a paginated list of posts.
func (s *PostService) GetPosts(ctx context.Context, page, pageSize int) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// normalize pagination params
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// get total count
	total, err := s.postRepo.CountPosts(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithContext(ctx).WithError(err).Error("GetPosts.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if total == 0 {
		httpResponse.Message = "no post found"
		httpStatusCode = http.StatusNotFound
		return
	}

	// get paginated posts
	posts, err := s.postRepo.ListPosts(ctx, pageSize, offset)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithContext(ctx).WithError(err).Error("GetPosts.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if len(posts) == 0 {
		httpResponse.Message = "no post found"
		httpStatusCode = http.StatusNotFound
		return
	}

	hasNext := int64(page*pageSize) < total
	hasPrevious := page > 1

	result := map[string]any{
		"hasNext":     hasNext,
		"hasPrevious": hasPrevious,
		"page":        page,
		"pageSize":    pageSize,
		"posts":       posts,
		"total":       total,
	}

	httpResponse.Message = result
	httpStatusCode = http.StatusOK
	return
}

// GetPost retrieves a post with the given postID.
func (s *PostService) GetPost(ctx context.Context, postID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	post, err := s.postRepo.GetPost(ctx, postID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "post not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithContext(ctx).WithError(err).Error("GetPost.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = post
	httpStatusCode = http.StatusOK
	return
}

// GetPostsByUserID retrieves all posts for a given userID.
func (s *PostService) GetPostsByUserID(ctx context.Context, userID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	posts, err := s.postRepo.GetPostsByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "no post found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithContext(ctx).WithError(err).Error("GetPostsByUserID.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = posts
	httpStatusCode = http.StatusOK
	return
}

// CreatePost creates a new post.
func (s *PostService) CreatePost(ctx context.Context, post *model.Post) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the user exists
	user, err := s.userRepo.GetUserByAuthID(ctx, post.IDAuth)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "no user profile found"
			httpStatusCode = http.StatusForbidden
			return
		}

		log.WithContext(ctx).WithError(err).Error("CreatePost.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	post.IDUser = user.UserID

	if err := s.postRepo.CreatePost(ctx, post); err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithContext(ctx).WithError(err).Error("CreatePost.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = post
	httpStatusCode = http.StatusCreated
	return
}

// UpdatePost updates an existing post.
func (s *PostService) UpdatePost(ctx context.Context, post *model.Post) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the post exists
	existingPost, err := s.postRepo.GetPost(ctx, post.PostID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "post not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithContext(ctx).WithError(err).Error("UpdatePost.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// ensure the post belongs to the user
	if existingPost.IDAuth != post.IDAuth {
		httpResponse.Message = "you are not allowed to update this post"
		httpStatusCode = http.StatusForbidden
		return
	}

	// if no changes are made, return the existing post
	if existingPost.Title == post.Title &&
		existingPost.Body == post.Body {
		httpResponse.Message = existingPost
		httpStatusCode = http.StatusOK
		return
	}

	// update the post fields
	existingPost.Title = post.Title
	existingPost.Body = post.Body

	// update the post
	if err := s.postRepo.UpdatePost(ctx, existingPost); err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithContext(ctx).WithError(err).Error("UpdatePost.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = existingPost
	httpStatusCode = http.StatusOK
	return
}

// DeletePost deletes a post with the given postID.
func (s *PostService) DeletePost(ctx context.Context, postID uint64, userIDAuth uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the post exists
	post, err := s.postRepo.GetPost(ctx, postID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "post not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithContext(ctx).WithError(err).Error("DeletePost.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// ensure the post belongs to the user
	if post.IDAuth != userIDAuth {
		httpResponse.Message = "you are not allowed to delete this post"
		httpStatusCode = http.StatusForbidden
		return
	}

	if err := s.postRepo.DeletePost(ctx, postID); err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithContext(ctx).WithError(err).Error("DeletePost.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = "post deleted successfully"
	httpStatusCode = http.StatusOK
	return
}

// DeletePostsByAuthID deletes all posts for a given authID.
func (s *PostService) DeletePostsByAuthID(ctx context.Context, authID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	user, err := s.userRepo.GetUserByAuthID(ctx, authID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "no user profile found"
			httpStatusCode = http.StatusForbidden
			return
		}

		log.WithContext(ctx).WithError(err).Error("DeletePostsByAuthID.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if err := s.postRepo.DeletePostsByUserID(ctx, user.UserID); err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithContext(ctx).WithError(err).Error("DeletePostsByAuthID.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = "all posts deleted successfully"
	httpStatusCode = http.StatusOK
	return
}
