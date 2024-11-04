package chat

import (
	"context"
	"time"
)

type PostDB struct {
	PostID        string       `bson:"post_id"`
	UserID        string       `bson:"user_id"`
	ForwardPostID string       `bson:"forward_post_id"`
	CommentPostID string       `bson:"comment_post_id"`
	RefPostID     string       `bson:"ref_post_id"`
	Content       string       `bson:"content"`
	AllowComment  int32        `bson:"allow_comment"`
	AllowForward  int32        `bson:"allow_forward"`
	AtUserIds     []string     `bson:"at_user_ids"`
	MediaMsgs     []*PostMedia `bson:"media_msgs"`
	CreateTime    time.Time    `bson:"create_time"`
	UpdateTime    time.Time    `bson:"update_time"`
}

type Post struct {
	PostID         string       `bson:"post_id"`
	ForwardPostID  string       `bson:"forward_post_id"`
	ForwardPost    *Post        `bson:"forward_post"`
	CommentPostID  string       `bson:"comment_post_id"`
	CommentPost    *Post        `bson:"comment_post"`
	RefPostID      string       `bson:"ref_post_id"`
	RefPost        *Post        `bson:"ref_post"`
	UserID         string       `bson:"user_id"`
	Content        string       `bson:"content"`
	AllowComment   int32        `bson:"allow_comment"`
	AllowForward   int32        `bson:"allow_forward"`
	AtUserIds      []string     `bson:"at_user_ids"`
	MediaMsgs      []*PostMedia `bson:"media_msgs"`
	CreateTime     time.Time    `bson:"create_time"`
	UpdateTime     time.Time    `bson:"update_time"`
	IsLiked        int32        `bson:"is_liked"`
	IsCollected    int32        `bson:"is_collected"`
	IsCommented    int32        `bson:"is_commented"`
	IsForwarded    int32        `bson:"is_forwarded"`
	CommentCount   int64        `bson:"comment_count"`
	LikeCount      int64        `bson:"like_count"`
	ForwardCount   int64        `bson:"forward_count"`
	UserInfo       *Attribute   `bson:"user_info"`
	AtUserInfoList []*Attribute `bson:"at_user_info_list"`
	IsPinned       int32        `bson:"is_pinned"`
}

type PostMedia struct {
	MediaType   int32       `bson:"media_type"`
	PostPicture PostPicture `bson:"post_picture,omitempty"`
	PostVideo   PostVideo   `bson:"post_video,omitempty"`
}

type PictureBaseInfo struct {
	UUID   string `bson:"uuid"`
	Type   string `bson:"type"`
	Size   int64  `bson:"size"`
	Width  int32  `bson:"width"`
	Height int32  `bson:"height"`
	URL    string `bson:"url"`
}

type PostPicture struct {
	SourcePath      string          `bson:"source_path"`
	SourcePicture   PictureBaseInfo `bson:"source_picture"`
	BigPicture      PictureBaseInfo `bson:"big_picture"`
	SnapshotPicture PictureBaseInfo `bson:"snapshot_picture"`
}

type PostVideo struct {
	VideoPath      string `bson:"video_path"`
	VideoUUID      string `bson:"video_uuid"`
	VideoURL       string `bson:"video_url"`
	VideoType      string `bson:"video_type"`
	VideoSize      int64  `bson:"video_size"`
	Duration       int64  `bson:"duration"`
	SnapshotPath   string `bson:"snapshot_path"`
	SnapshotUUID   string `bson:"snapshot_uuid"`
	SnapshotSize   int64  `bson:"snapshot_size"`
	SnapshotURL    string `bson:"snapshot_url"`
	SnapshotWidth  int32  `bson:"snapshot_width"`
	SnapshotHeight int32  `bson:"snapshot_height"`
	SnapshotType   string `bson:"snapshot_type"`
}

func (Post) TableName() string {
	return "posts"
}

type PostInterface interface {
	// 创建帖子
	Create(ctx context.Context, posts []*PostDB) error
	// 通过帖子ID获取帖子
	Take(ctx context.Context, postID string) (*Post, error)
	// 更新帖子
	UpdateByMap(ctx context.Context, postID string, data map[string]any) error
	// 删除帖子
	Delete(ctx context.Context, postIDs []string) error
	// 通过转发的帖子ID获取帖子
	GetPostByForwardPostID(ctx context.Context, userID, forwardPostID string) (*Post, error)
	// 通过游标和用户IDs获取此ID后Count数的帖子
	GetPostsByCursorAndUserIDs(ctx context.Context, cursor int64, userIDs []string, count int64) ([]*Post, string, error)
	// 通过游标和用户ID获取此ID后Count数的帖子
	GetPostsByCursorAndUser(ctx context.Context, cursor int64, userID string, count int64) ([]*Post, string, error)
	// 通过游标和帖子IDs获取此ID后Count数的帖子
	GetPostsByCursorAndPostIDs(ctx context.Context, cursor int64, postIDs []string, count int64) ([]*Post, string, error)
	// 通过游标和帖子ID获取评论帖子
	GetCommentPostsByPostID(ctx context.Context, cursor int64, postID string, count int64) ([]*Post, string, error)
	// 获取自身评论的帖子，和被评论的帖子 IDs
	GetCommentPostIDsByUser(ctx context.Context, userID string) ([]string, error)
	// 获取关注用户的IDs
	GetFollowedUserIDs(ctx context.Context, userID string) ([]string, error)
	// 获取订阅用户的IDs
	GetSubscriberUserIDs(ctx context.Context, userID string) ([]string, error)
	// 获取置顶帖子
	GetPinnedPostByUserID(ctx context.Context, userID string) (*Post, error)
}
