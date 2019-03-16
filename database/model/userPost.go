package model

// UserPost model - intermediate table `user_posts` (many to many relations)
type UserPost struct {
	UserID uint
	PostID uint
}
