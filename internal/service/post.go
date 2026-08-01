package service

import (
	"errors"
	"net/http"
	"time"
	"unicode/utf8"

	"forum/internal/apperror"
	"forum/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostService struct {
	db *gorm.DB
}

func NewPostService(db *gorm.DB) *PostService {
	return &PostService{db: db}
}

type PostCreatedView struct {
	ID        uint64           `json:"id"`
	Content   string           `json:"content"`
	Author    model.AuthorView `json:"author"`
	CreatedAt time.Time        `json:"created_at"`
}

type PostItemView struct {
	ID           uint64           `json:"id"`
	Content      string           `json:"content"`
	Author       model.AuthorView `json:"author"`
	LikeCount    int64            `json:"like_count"`
	CommentCount int64            `json:"comment_count"`
	CreatedAt    time.Time        `json:"created_at"`
}

type PageMeta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type PostListView struct {
	Items []PostItemView `json:"items"`
	Meta  PageMeta       `json:"meta"`
}

type CommentView struct {
	ID        uint64           `json:"id"`
	PostID    uint64           `json:"post_id"`
	Content   string           `json:"content"`
	Author    model.AuthorView `json:"author"`
	CreatedAt time.Time        `json:"created_at"`
}

type PostDetailView struct {
	ID           uint64           `json:"id"`
	Content      string           `json:"content"`
	Author       model.AuthorView `json:"author"`
	LikeCount    int64            `json:"like_count"`
	CommentCount int64            `json:"comment_count"`
	CreatedAt    time.Time        `json:"created_at"`
	Comments     []CommentView    `json:"comments"`
}

func (s *PostService) Create(userID uint64, content string) (PostCreatedView, error) {
	if !validPostContent(content) {
		return PostCreatedView{}, apperror.ErrBadRequest
	}
	var author model.User
	if err := s.db.First(&author, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return PostCreatedView{}, apperror.ErrUnauthorized
		}
		return PostCreatedView{}, err
	}
	post := model.Post{UserID: userID, Content: content}
	if err := s.db.Create(&post).Error; err != nil {
		return PostCreatedView{}, err
	}
	return PostCreatedView{
		ID:        post.ID,
		Content:   post.Content,
		Author:    author.AuthorView(),
		CreatedAt: post.CreatedAt,
	}, nil
}

func (s *PostService) List(page, pageSize int, sort string) (PostListView, error) {
	var total int64
	var posts []model.Post
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Post{}).Count(&total).Error; err != nil {
			return err
		}
		query := tx.Preload("Author")
		//两种排序
		if sort == "hot" {
			query = query.Order("(like_count * 2 + comment_count * 3) DESC").Order("created_at DESC").Order("id DESC")
		} else {
			query = query.Order("created_at DESC").Order("id DESC")
		}
		return query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&posts).Error
	})
	if err != nil {
		return PostListView{}, err
	}
	items := make([]PostItemView, 0, len(posts))
	for _, post := range posts {
		items = append(items, toPostItem(post))
	}
	return PostListView{Items: items, Meta: PageMeta{Page: page, PageSize: pageSize, Total: total}}, nil
}

func (s *PostService) Detail(postID uint64) (PostDetailView, error) {
	var post model.Post
	var comments []model.Comment
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Author").First(&post, postID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.New(http.StatusNotFound, http.StatusNotFound, "帖子不存在")
			}
			return err
		}
		return tx.Preload("Author").Where("post_id = ?", postID).Order("created_at ASC").Order("id ASC").Find(&comments).Error
	})
	if err != nil {
		return PostDetailView{}, err
	}
	commentViews := make([]CommentView, 0, len(comments))
	for _, comment := range comments {
		commentViews = append(commentViews, toCommentView(comment))
	}
	return PostDetailView{
		ID:           post.ID,
		Content:      post.Content,
		Author:       post.Author.AuthorView(),
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		CreatedAt:    post.CreatedAt,
		Comments:     commentViews,
	}, nil
}

func (s *PostService) DeleteOwn(userID, postID uint64) error {
	return deletePostRecords(s.db, postID, &userID)
}

func (s *PostService) DeleteAny(postID uint64) error {
	return deletePostRecords(s.db, postID, nil)
}

func deletePostRecords(db *gorm.DB, postID uint64, ownerID *uint64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var post model.Post
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&post, postID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.New(http.StatusNotFound, http.StatusNotFound, "帖子不存在")
			}
			return err
		}
		if ownerID != nil && post.UserID != *ownerID {
			return apperror.New(http.StatusForbidden, http.StatusForbidden, "无权删除他人的帖子")
		}
		if err := tx.Where("post_id = ?", postID).Delete(&model.PostLike{}).Error; err != nil {
			return err
		}
		if err := tx.Where("post_id = ?", postID).Delete(&model.Comment{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&model.Post{}, postID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return apperror.New(http.StatusNotFound, http.StatusNotFound, "帖子不存在")
		}
		return nil
	})
}

func validPostContent(content string) bool {
	length := utf8.RuneCountInString(content)
	return length >= 1 && length <= 2000
}

func toPostItem(post model.Post) PostItemView {
	return PostItemView{
		ID:           post.ID,
		Content:      post.Content,
		Author:       post.Author.AuthorView(),
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		CreatedAt:    post.CreatedAt,
	}
}

func toCommentView(comment model.Comment) CommentView {
	return CommentView{
		ID:        comment.ID,
		PostID:    comment.PostID,
		Content:   comment.Content,
		Author:    comment.Author.AuthorView(),
		CreatedAt: comment.CreatedAt,
	}
}
