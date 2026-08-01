package service

import (
	"context"
	"errors"
	"net/http"

	"forum/internal/apperror"
	"forum/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SocialService struct {
	db *gorm.DB
}

func NewSocialService(db *gorm.DB) *SocialService {
	return &SocialService{db: db}
}

type LikeResult struct {
	PostID  uint64 `json:"post_id"`
	IsLiked bool   `json:"is_liked"`
}

type LikeStatus struct {
	PostID uint64 `json:"post_id"`
	Liked  bool   `json:"liked"`
}

type LikeStatusesResult struct {
	Status []LikeStatus `json:"status"`
}
//点赞*取消点赞
func (s *SocialService) ToggleLike(ctx context.Context, userID, postID uint64) (LikeResult, error) {
	liked := false
	db := s.db.WithContext(ctx)
	err := db.Transaction(func(tx *gorm.DB) error {
		var post model.Post
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&post, postID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.New(http.StatusNotFound, http.StatusNotFound, "帖子不存在")
			}
			return err
		}

		var like model.PostLike
		err := tx.Where("post_id = ? AND user_id = ?", postID, userID).First(&like).Error
		switch {
		case err == nil:
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Post{}).Where("id = ?", postID).Updates(map[string]any{
				"like_count":   gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END"),
				"like_version": gorm.Expr("like_version + 1"),
			}).Error; err != nil {
				return err
			}
			liked = false
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := tx.Create(&model.PostLike{PostID: postID, UserID: userID}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Post{}).Where("id = ?", postID).Updates(map[string]any{
				"like_count":   gorm.Expr("like_count + 1"),
				"like_version": gorm.Expr("like_version + 1"),
			}).Error; err != nil {
				return err
			}
			liked = true
		default:
			return err
		}
		return nil
	})
	if err != nil {
		return LikeResult{}, err
	}
	return LikeResult{PostID: postID, IsLiked: liked}, nil
}

func (s *SocialService) LikeStatuses(ctx context.Context, userID uint64, postIDs []uint64) (LikeStatusesResult, error) {
	likedIDs := make(map[uint64]struct{}, len(postIDs))
	var likes []model.PostLike
	if err := s.db.WithContext(ctx).Select("post_id").Where("user_id = ? AND post_id IN ?", userID, postIDs).Find(&likes).Error; err != nil {
		return LikeStatusesResult{}, err
	}
	for _, like := range likes {
		likedIDs[like.PostID] = struct{}{}
	}

	statuses := make([]LikeStatus, 0, len(postIDs))
	for _, postID := range postIDs {
		_, liked := likedIDs[postID]
		statuses = append(statuses, LikeStatus{PostID: postID, Liked: liked})
	}
	return LikeStatusesResult{Status: statuses}, nil
}
//评论
func (s *SocialService) AddComment(ctx context.Context, userID, postID uint64, content string) (CommentView, error) {
	var result CommentView
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var post model.Post
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&post, postID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.New(http.StatusNotFound, http.StatusNotFound, "帖子不存在")
			}
			return err
		}
		var author model.User
		if err := tx.First(&author, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.ErrUnauthorized
			}
			return err
		}
		comment := model.Comment{PostID: postID, UserID: userID, Content: content}
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Post{}).Where("id = ?", postID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error; err != nil {
			return err
		}
		comment.Author = author
		result = toCommentView(comment)
		return nil
	})
	return result, err
}
