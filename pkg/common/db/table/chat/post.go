package chat

import (
	"context"
	"time"

	"github.com/openimsdk/tools/db/pagination"
)

type PostDB struct {
	PostID        string       `bson:"post_id"`
	ForwardPostID string       `bson:"forward_post_id"`
	UserID        string       `bson:"user_id"`
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
	IsForwarded    int32        `bson:"is_forwarded"`
	CommentCount   int64        `bson:"comment_count"`
	LikeCount      int64        `bson:"like_count"`
	CollectCount   int64        `bson:"collect_count"`
	ForwardCount   int64        `bson:"forward_count"`
	UserInfo       *Attribute   `bson:"user_info"`
	AtUserInfoList []*Attribute `bson:"at_user_info_list"`
}

type PostMedia struct {
	MediaType   int32       `bson:"media_type"`
	PostPicture PostPicture `bson:"post_picture"`
	PostVideo   PostVideo   `bson:"post_video"`
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
	Create(ctx context.Context, posts []*PostDB) error
	Take(ctx context.Context, postID string) (*Post, error)
	UpdateByMap(ctx context.Context, postID string, data map[string]any) error
	Delete(ctx context.Context, postID string) error
	PageGet(ctx context.Context, pagination pagination.Pagination) (int64, []*Post, error)
	PageGetByUser(ctx context.Context, userID string, pagination pagination.Pagination) (int64, []*Post, error)
	PageGetByPostIDs(ctx context.Context, postIDs []string, pagination pagination.Pagination) (int64, []*Post, error)
}
