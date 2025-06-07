package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	grenderer "github.com/pilinux/gorest/lib/renderer"

	"github.com/pilinux/gorest/example2/internal/database/model"
	"github.com/pilinux/gorest/example2/internal/service"
)

// PostAPI provides HTTP handlers for post-related endpoints.
type PostAPI struct {
	postService *service.PostService
}

// NewPostAPI returns a new PostAPI instance.
func NewPostAPI(postService *service.PostService) *PostAPI {
	return &PostAPI{
		postService: postService,
	}
}

// GetPosts handles the HTTP GET request to retrieve all posts.
//
// Endpoint: GET /api/v1/posts
//
// Authorization: None
func (api *PostAPI) GetPosts(c *gin.Context) {
	resp, statusCode := api.postService.GetPosts()
	grenderer.Render(c, resp, statusCode)
}

// GetPost handles the HTTP GET request to retrieve a post by ID.
//
// Endpoint: GET /api/v1/posts/:id
//
// Authorization: None
func (api *PostAPI) GetPost(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		grenderer.Render(c, "postID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		grenderer.Render(c, "invalid postID format", http.StatusBadRequest)
		return
	}

	resp, statusCode := api.postService.GetPost(postID)
	grenderer.Render(c, resp, statusCode)
}

// CreatePost handles the HTTP POST request to create a new post.
//
// Endpoint: POST /api/v1/posts
//
// Authorization: JWT token required
func (api *PostAPI) CreatePost(c *gin.Context) {
	userIDAuth := getAuthID(c)
	if userIDAuth == 0 {
		grenderer.Render(c, "unauthorized", http.StatusUnauthorized)
		return
	}

	var post model.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	if err := post.Trim(); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	post.IDAuth = userIDAuth

	resp, statusCode := api.postService.CreatePost(&post)
	grenderer.Render(c, resp, statusCode)
}

// UpdatePost handles the HTTP PUT request to update an existing post.
//
// Endpoint: PUT /api/v1/posts/:id
//
// Authorization: JWT token required
func (api *PostAPI) UpdatePost(c *gin.Context) {
	userIDAuth := getAuthID(c)
	if userIDAuth == 0 {
		grenderer.Render(c, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		grenderer.Render(c, "postID is required", http.StatusBadRequest)
		return
	}
	postID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		grenderer.Render(c, "invalid postID format", http.StatusBadRequest)
		return
	}

	var post model.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	if err := post.Trim(); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	post.PostID = postID
	post.IDAuth = userIDAuth

	resp, statusCode := api.postService.UpdatePost(&post)
	grenderer.Render(c, resp, statusCode)
}

// DeletePost handles the HTTP DELETE request to delete a post.
//
// Endpoint: DELETE /api/v1/posts/:id
//
// Authorization: JWT token required
func (api *PostAPI) DeletePost(c *gin.Context) {
	userIDAuth := getAuthID(c)
	if userIDAuth == 0 {
		grenderer.Render(c, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		grenderer.Render(c, "postID is required", http.StatusBadRequest)
		return
	}
	postID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		grenderer.Render(c, "invalid postID format", http.StatusBadRequest)
		return
	}

	resp, statusCode := api.postService.DeletePost(postID, userIDAuth)
	grenderer.Render(c, resp, statusCode)
}

// DeleteAllPostsOfUser handles the HTTP DELETE request to delete all posts of a user.
//
// Endpoint: DELETE /api/v1/posts/all
//
// Authorization: JWT token required
func (api *PostAPI) DeleteAllPostsOfUser(c *gin.Context) {
	userIDAuth := getAuthID(c)
	if userIDAuth == 0 {
		grenderer.Render(c, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, statusCode := api.postService.DeletePostsByAuthID(userIDAuth)
	grenderer.Render(c, resp, statusCode)
}
