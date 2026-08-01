package model

import "time"

const (
	RoleStudent = "student"
	RoleAdmin   = "admin"
)

type User struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Name         string    `gorm:"size:64;not null" json:"name"`
	Role         string    `gorm:"size:16;not null;default:student" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"-"`
}

type Post struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	UserID       uint64    `gorm:"index;not null" json:"-"`
	Author       User      `gorm:"foreignKey:UserID" json:"author"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	LikeCount    int64     `gorm:"not null;default:0" json:"like_count"`
	LikeVersion  uint64    `gorm:"not null;default:0" json:"-"`
	CommentCount int64     `gorm:"not null;default:0" json:"comment_count"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Comment struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	PostID    uint64    `gorm:"index;not null" json:"post_id"`
	UserID    uint64    `gorm:"index;not null" json:"-"`
	Author    User      `gorm:"foreignKey:UserID" json:"author"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

type PostLike struct {
	ID        uint64    `gorm:"primaryKey" json:"-"`
	PostID    uint64    `gorm:"uniqueIndex:idx_post_user;not null" json:"post_id"`
	UserID    uint64    `gorm:"uniqueIndex:idx_post_user;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthorView struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

func (u User) AuthorView() AuthorView {
	return AuthorView{ID: u.ID, Username: u.Username, Name: u.Name, Role: u.Role}
}
