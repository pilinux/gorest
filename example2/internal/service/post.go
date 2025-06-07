package service

import (
	"net/http"

	gmodel "github.com/pilinux/gorest/database/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/pilinux/gorest/example2/internal/database/model"
	"github.com/pilinux/gorest/example2/internal/repo"
)

// PostService provides methods for post-related operations.
type PostService struct {
	postRepo *repo.PostRepo
	userRepo *repo.UserRepo
}

// NewPostService returns a new PostService instance.
func NewPostService(postRepo *repo.PostRepo, userRepo *repo.UserRepo) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// GetPosts retrieves all posts.
func (s *PostService) GetPosts() (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	posts, err := s.postRepo.GetPosts()
	if err != nil {
		log.WithError(err).Error("GetPosts.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if len(posts) == 0 {
		httpResponse.Message = "no post found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = posts
	httpStatusCode = http.StatusOK
	return
}

// GetPost retrieves a post with the given postID.
func (s *PostService) GetPost(postID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	post, err := s.postRepo.GetPost(postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpResponse.Message = "post not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("GetPost.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = post
	httpStatusCode = http.StatusOK
	return
}

// GetPostsByUserID retrieves all posts for a given userID.
func (s *PostService) GetPostsByUserID(userID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	posts, err := s.postRepo.GetPostsByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpResponse.Message = "no post found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("GetPostsByUserID.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = posts
	httpStatusCode = http.StatusOK
	return
}

// CreatePost creates a new post.
func (s *PostService) CreatePost(post *model.Post) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the user exists
	user, err := s.userRepo.GetUserByAuthID(post.IDAuth)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpResponse.Message = "no user profile found"
			httpStatusCode = http.StatusForbidden
			return
		}

		log.WithError(err).Error("CreatePost.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	post.IDUser = user.UserID

	if err := s.postRepo.CreatePost(post); err != nil {
		log.WithError(err).Error("CreatePost.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = post
	httpStatusCode = http.StatusCreated
	return
}

// UpdatePost updates an existing post.
func (s *PostService) UpdatePost(post *model.Post) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the post exists
	existingPost, err := s.postRepo.GetPost(post.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpResponse.Message = "post not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("UpdatePost.s.1")
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
	if err := s.postRepo.UpdatePost(existingPost); err != nil {
		log.WithError(err).Error("UpdatePost.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = existingPost
	httpStatusCode = http.StatusOK
	return
}

// DeletePost deletes a post with the given postID.
func (s *PostService) DeletePost(postID uint64, userIDAuth uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the post exists
	post, err := s.postRepo.GetPost(postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpResponse.Message = "post not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("DeletePost.s.1")
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

	if err := s.postRepo.DeletePost(postID); err != nil {
		log.WithError(err).Error("DeletePost.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = "post deleted successfully"
	httpStatusCode = http.StatusOK
	return
}

// DeletePostsByAuthID deletes all posts for a given authID.
func (s *PostService) DeletePostsByAuthID(authID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	user, err := s.userRepo.GetUserByAuthID(authID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpResponse.Message = "no user profile found"
			httpStatusCode = http.StatusForbidden
			return
		}

		log.WithError(err).Error("DeletePostsByAuthID.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if err := s.postRepo.DeletePostsByUserID(user.UserID); err != nil {
		log.WithError(err).Error("DeletePostsByAuthID.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = "all posts deleted successfully"
	httpStatusCode = http.StatusOK
	return
}
