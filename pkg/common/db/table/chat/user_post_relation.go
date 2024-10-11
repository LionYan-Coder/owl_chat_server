package chat

import (
	"context"
	"time"
)

type UserPostRelation struct {
	RelationID   string    `bson:"relation_id"`
	UserID       string    `bson:"user_id"`
	PostID       string    `bson:"post_id"`
	IsFollowed   int32     `bson:"is_followed"`
	IsSubscribed int32     `bson:"is_subscribed"`
	IsLiked      int32     `bson:"is_liked"`
	IsCollected  int32     `bson:"is_collected"`
	IsForwarded  int32     `bson:"is_forwarded"`
	IsCommented  int32     `bson:"is_commented"`
	CreateTime   time.Time `bson:"create_time"`
	UpdateTime   time.Time `bson:"update_time"`
}

func (UserPostRelation) TableName() string {
	return "user_post_relations"
}

type UserPostRelationInterface interface {
	Create(ctx context.Context, posts []*UserPostRelation) error
	Take(ctx context.Context, userID, postID string) (*UserPostRelation, error)
	UpdateByMap(ctx context.Context, userID, postID string, args map[string]any) error
	Delete(ctx context.Context, userID, postID string) error
	GetLikeCount(ctx context.Context, postID string) (int64, error)
	GetCollectCount(ctx context.Context, postID string) (int64, error)
	GetForwardCount(ctx context.Context, postID string) (int64, error)
	GetCommentCount(ctx context.Context, postID string) (int64, error)
	GetFollowCount(ctx context.Context, userID string) (int64, error)
	GetSubscribeCount(ctx context.Context, userID string) (int64, error)
	GetLikeCounts(ctx context.Context, postIDs []string) (map[string]int64, error)
	GetCommentCounts(ctx context.Context, postIDs []string) (map[string]int64, error)
	GetForwardCounts(ctx context.Context, postIDs []string) (map[string]int64, error)
	GetIsLiked(ctx context.Context, userID, postID string) (int32, error)
	GetIsCollected(ctx context.Context, userID, postID string) (int32, error)
	GetIsForwarded(ctx context.Context, userID, postID string) (int32, error)
	GetIsCommented(ctx context.Context, userID, postID string) (int32, error)
	GetFollowPostIDs(ctx context.Context, userID string) ([]string, error)
	GetSubscribePostIDs(ctx context.Context, userID string) ([]string, error)
	GetReplyPostIDs(ctx context.Context, userID string) ([]string, error)
	GetLikedPostIDs(ctx context.Context, userID string) ([]string, error)
	GetCollectedPostIDs(ctx context.Context, userID string) ([]string, error)
}
