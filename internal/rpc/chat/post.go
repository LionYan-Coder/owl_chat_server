package chat

import (
	"context"
	"errors"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/openimsdk/chat/pkg/redpacket/servererrs"
	"github.com/openimsdk/tools/db/tx"
	"github.com/openimsdk/tools/mcontext"
	"github.com/openimsdk/tools/utils/encrypt"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/openimsdk/chat/pkg/common/constant"
	"github.com/openimsdk/chat/pkg/common/convert"
	"github.com/openimsdk/chat/pkg/common/db/table/chat"
	"github.com/openimsdk/chat/pkg/common/mctx"
	chatpb "github.com/openimsdk/chat/pkg/protocol/chat"
	"github.com/openimsdk/tools/errs"
)

func (o *chatSvr) PublishPost(ctx context.Context, req *chatpb.PublishPostReq) (*chatpb.PublishPostResp, error) {
	userID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}
	postDB := &chat.PostDB{
		UserID:       userID,
		AllowComment: req.AllowComment,
		AllowForward: req.AllowForward,
		Content:      req.Content.Value,
		AtUserIds:    req.AtUserIds,
		MediaMsgs:    convert.PostMediasPb2DB(req.MediaMsgs),
	}
	if err := o.GenPostID(ctx, &postDB.PostID); err != nil {
		return nil, err
	}
	err = o.Database.CreatePost(ctx, []*chat.PostDB{postDB})
	if err != nil {
		return nil, err
	}
	post, err := o.Database.GetPostByID(ctx, postDB.PostID)
	if err != nil {
		return nil, err
	}
	return &chatpb.PublishPostResp{
		Post: convert.PostDB2Pb(post),
	}, nil
}

func (o *chatSvr) ForwardPost(ctx context.Context, req *chatpb.ForwardPostReq) (*chatpb.ForwardPostResp, error) {
	userID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}

	err = o.tx.Transaction(ctx, func(ctx context.Context) error {

		if req.IsForwarded == constant.Forwarded {
			postDB := &chat.PostDB{
				UserID:        userID,
				ForwardPostID: req.ForwardPostID,
			}
			if err := o.GenPostID(ctx, &postDB.PostID); err != nil {
				return err
			}
			if err := o.Database.CreatePost(ctx, []*chat.PostDB{postDB}); err != nil {
				return err
			}

			relation, err := o.Database.GetUserPostRelation(ctx, userID, req.ForwardPostID)
			switch {
			case err == nil:
				relation.IsForwarded = constant.Forwarded
				err = o.Database.UpdateUserPostRelation(ctx, userID, req.ForwardPostID, map[string]any{"is_forwarded": relation.IsForwarded})
				if err != nil {
					return err
				}
			case errors.Is(err, mongo.ErrNoDocuments):
				relation = &chat.UserPostRelation{
					UserID:      userID,
					PostID:      req.ForwardPostID,
					IsForwarded: constant.Forwarded,
				}
				err = o.Database.CreateUserPostRelation(ctx, []*chat.UserPostRelation{relation})
				if err != nil {
					return err
				}
			default:
				return err
			}

		} else {
			if err := o.Database.UpdateUserPostRelation(ctx, userID, req.ForwardPostID, map[string]any{"is_forwarded": constant.NotForwarded}); err != nil {
				return err
			}
			post, err := o.Database.GetPostByForwardPostID(ctx, userID, req.ForwardPostID)
			if err != nil {
				return err
			}
			if err := o.Database.DeletePost(ctx, []string{post.PostID}); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &chatpb.ForwardPostResp{
		IsForwarded: req.IsForwarded,
	}, nil
}

func (o *chatSvr) CommentPost(ctx context.Context, req *chatpb.CommentPostReq) (*chatpb.CommentPostResp, error) {
	userID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}
	postDB := &chat.PostDB{
		UserID:        userID,
		CommentPostID: req.CommentPostID,
		AllowComment:  req.AllowComment,
		AllowForward:  req.AllowForward,
		Content:       req.Content.Value,
		AtUserIds:     req.AtUserIds,
		MediaMsgs:     convert.PostMediasPb2DB(req.MediaMsgs),
	}
	if err := o.GenPostID(ctx, &postDB.PostID); err != nil {
		return nil, err
	}
	relation, err := o.Database.GetUserPostRelation(ctx, userID, req.CommentPostID)
	switch {
	case err == nil:
		relation.IsCommented = constant.Commented
		err = o.Database.UpdateUserPostRelation(ctx, userID, req.CommentPostID, map[string]any{"is_commented": relation.IsCommented})
		if err != nil {
			return nil, err
		}
	case errors.Is(err, mongo.ErrNoDocuments):
		err = tx.Tx.Transaction(o.tx, ctx, func(ctx context.Context) error {
			relation = &chat.UserPostRelation{
				UserID:      userID,
				PostID:      req.CommentPostID,
				IsCommented: constant.Commented,
			}
			if err := o.Database.CreatePost(ctx, []*chat.PostDB{postDB}); err != nil {
				return err
			}
			if err := o.Database.CreateUserPostRelation(ctx, []*chat.UserPostRelation{relation}); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return nil, err
		}

	default:
		return nil, err
	}
	post, err := o.Database.GetPostByID(ctx, postDB.PostID)
	if err != nil {
		return nil, err
	}
	return &chatpb.CommentPostResp{
		Post: convert.PostDB2Pb(post),
	}, nil
}

func (o *chatSvr) PinPost(ctx context.Context, req *chatpb.PinPostReq) (*chatpb.PinPostResp, error) {
	userID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}

	post, err := o.Database.GetPostByID(ctx, req.PostID)
	if err != nil {
		return nil, err
	}

	if post.UserID != userID {
		return nil, errs.ErrNoPermission.WrapMsg("permission denied")
	}

	err = o.tx.Transaction(ctx, func(ctx context.Context) error {
		if req.IsPinned == constant.Pinned {
			pinnedPost, err := o.Database.GetPinnedPostByUserID(ctx, userID)
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return err
			}
			if pinnedPost != nil {
				err = o.Database.UpdatePost(ctx, pinnedPost.PostID, map[string]any{"is_pined": constant.UnPinned})
				if err != nil {
					return err
				}
			}
		}
		err = o.Database.UpdatePost(ctx, req.PostID, map[string]any{"is_pined": req.IsPinned})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &chatpb.PinPostResp{}, nil
}

func (o *chatSvr) ReferencePost(ctx context.Context, req *chatpb.ReferencePostReq) (*chatpb.ReferencePostResp, error) {
	userID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}
	postDB := &chat.PostDB{
		UserID:       userID,
		RefPostID:    req.RefPostID,
		Content:      req.Content.Value,
		AllowComment: req.AllowComment,
		AllowForward: req.AllowForward,
		AtUserIds:    req.AtUserIds,
		MediaMsgs:    convert.PostMediasPb2DB(req.MediaMsgs),
	}
	if err := o.GenPostID(ctx, &postDB.PostID); err != nil {
		return nil, err
	}
	err = o.Database.CreatePost(ctx, []*chat.PostDB{postDB})
	if err != nil {
		return nil, err
	}
	return &chatpb.ReferencePostResp{}, nil
}

func (o *chatSvr) ChangeLikePost(ctx context.Context, req *chatpb.LikePostReq) (*chatpb.LikePostResp, error) {
	opUserID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}
	post, err := o.Database.GetPostByID(ctx, req.PostID)
	if err != nil {
		return nil, err
	}

	isLiked := constant.NotLiked
	if req.IsLiked == constant.Liked {
		isLiked = constant.Liked
	}

	relation, err := o.Database.GetUserPostRelation(ctx, opUserID, post.PostID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			relation = &chat.UserPostRelation{
				UserID:  opUserID,
				PostID:  post.PostID,
				IsLiked: int32(isLiked),
			}
			err = o.Database.CreateUserPostRelation(ctx, []*chat.UserPostRelation{relation})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}

	} else {
		relation.IsLiked = int32(isLiked)
		err = o.Database.UpdateUserPostRelation(ctx, opUserID, post.PostID, map[string]any{"is_liked": relation.IsLiked})
		if err != nil {
			return nil, err
		}
	}

	return &chatpb.LikePostResp{
		IsLiked: relation.IsLiked,
	}, nil
}

func (o *chatSvr) ChangeCollectPost(ctx context.Context, req *chatpb.CollectPostReq) (*chatpb.CollectPostResp, error) {
	opUserID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}
	post, err := o.Database.GetPostByID(ctx, req.PostID)
	if err != nil {
		return nil, err
	}

	isCollected := constant.NotCollected
	if req.IsCollected == constant.Collected {
		isCollected = constant.Collected
	}

	relation, err := o.Database.GetUserPostRelation(ctx, opUserID, post.PostID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			relation = &chat.UserPostRelation{
				UserID:      opUserID,
				PostID:      post.PostID,
				IsCollected: int32(isCollected),
			}
			err = o.Database.CreateUserPostRelation(ctx, []*chat.UserPostRelation{relation})
			if err != nil {
				return nil, err
			}

		} else {
			return nil, err
		}

	} else {
		relation.IsCollected = int32(isCollected)
		err = o.Database.UpdateUserPostRelation(ctx, opUserID, post.PostID, map[string]any{"is_collected": relation.IsCollected})
		if err != nil {
			return nil, err
		}
	}

	return &chatpb.CollectPostResp{
		IsCollected: relation.IsCollected,
	}, nil
}

func (o *chatSvr) ChangeAllowCommentPost(ctx context.Context, req *chatpb.ChangeAllowCommentPostReq) (*chatpb.ChangeAllowCommentPostResp, error) {
	opUserID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}
	post, err := o.Database.GetPostByID(ctx, req.PostID)
	if err != nil {
		return nil, err
	}
	if post.UserID != opUserID {
		return nil, errs.ErrNoPermission.WrapMsg("permission denied")
	}

	allowComment := constant.NotCommented
	if post.AllowComment == constant.NotCommented {
		allowComment = constant.Commented
	}

	err = o.Database.UpdatePost(ctx, req.PostID, map[string]any{"allow_comment": allowComment})
	if err != nil {
		return nil, err
	}

	return &chatpb.ChangeAllowCommentPostResp{
		PostID:       req.PostID,
		AllowComment: int32(allowComment),
	}, nil
}

func (o *chatSvr) ChangeAllowForwardPost(ctx context.Context, req *chatpb.ChangeAllowForwardPostReq) (*chatpb.ChangeAllowForwardPostResp, error) {
	opUserID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}
	post, err := o.Database.GetPostByID(ctx, req.PostID)
	if err != nil {
		return nil, err
	}
	if post.UserID != opUserID {
		return nil, errs.ErrNoPermission.WrapMsg("permission denied")
	}

	allowForward := constant.NotForwarded
	if post.AllowForward == constant.NotForwarded {
		allowForward = constant.Forwarded
	}

	err = o.Database.UpdatePost(ctx, req.PostID, map[string]any{"allow_forward": allowForward})
	if err != nil {
		return nil, err
	}

	return &chatpb.ChangeAllowForwardPostResp{
		PostID:       req.PostID,
		AllowForward: int32(allowForward),
	}, nil
}

func (o *chatSvr) DeletePost(ctx context.Context, req *chatpb.DeletePostReq) (*chatpb.DeletePostResp, error) {
	opUserID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}
	post, err := o.Database.GetPostByID(ctx, req.PostID)
	if err != nil {
		return nil, err
	}
	if post.UserID != opUserID {
		return nil, errs.ErrNoPermission.WrapMsg("permission denied")
	}

	err = o.Database.DeletePost(ctx, []string{req.PostID})
	if err != nil {
		return nil, err
	}

	return &chatpb.DeletePostResp{}, nil
}

func (o *chatSvr) GetPostByID(ctx context.Context, req *chatpb.GetPostByIDReq) (*chatpb.GetPostByIDResp, error) {
	post, err := o.Database.GetPostByID(ctx, req.PostID)
	if err != nil {
		return nil, err
	}
	postPB := convert.PostDB2Pb(post)
	return &chatpb.GetPostByIDResp{
		Post: postPB,
	}, nil
}

func (o *chatSvr) GetPostListByUser(ctx context.Context, req *chatpb.GetPostListByUserReq) (*chatpb.GetPostListByUserResp, error) {
	resp := &chatpb.GetPostListByUserResp{}
	postsDB, nextCursor, err := o.Database.GetPostsByCursorAndUser(ctx, req.NextCursor, req.UserID, int64(req.Count))
	if err != nil {
		return nil, err
	}
	postsPB := convert.PostsDB2Pb(postsDB)

	resp.Posts = postsPB
	if nextCursor != "" {
		nextCursorInt, err := strconv.ParseInt(nextCursor, 10, 64)
		if err != nil {
			return nil, err
		}
		resp.NextCursor = nextCursorInt
	}
	return resp, nil
}

func (o *chatSvr) GetAllTypePost(ctx context.Context, req *chatpb.GetAllTypePostReq) (*chatpb.GetAllTypePostResp, error) {
	resp := &chatpb.GetAllTypePostResp{}
	var allPosts []*chatpb.TypePosts
	postTypes := []int32{constant.Follow, constant.Subscribe, constant.Reply, constant.Like, constant.Collect}
	for _, postType := range postTypes {
		paginationReq := &chatpb.GetPostListReq{
			Type:  postType,
			Count: req.Count,
		}

		paginationResp, err := o.GetPostList(ctx, paginationReq)
		if err != nil {
			return nil, err
		}
		typePosts := &chatpb.TypePosts{
			Type:       postType,
			NextCursor: paginationResp.NextCursor,
			Posts:      paginationResp.Posts,
		}
		allPosts = append(allPosts, typePosts)
	}
	resp.AllPosts = allPosts
	return resp, nil
}

func (o *chatSvr) GetPostList(ctx context.Context, req *chatpb.GetPostListReq) (*chatpb.GetPostListResp, error) {
	userID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}
	resp := &chatpb.GetPostListResp{}

	var cursor = req.NextCursor
	var nextCursor string
	var postsDB []*chat.Post

	/// 根据不同的类型获取不同的帖子
	/// 关注 获取关注人的帖子
	/// 订阅 获取订阅人的帖子
	/// 回复 获取帖子被回复的帖子 （除了这个类型，其他不返回回复的帖子）
	/// 点赞 获取帖子被点赞的帖子
	/// 收藏 获取帖子被收藏的帖子

	switch req.Type {
	case constant.Follow:
		userIDs, err := o.Database.GetFollowedUserIDs(ctx, userID)
		if err != nil {
			return nil, err
		}
		userIDs = append([]string{userID}, userIDs...)
		postsDB, nextCursor, err = o.Database.GetPostsByCursorAndUserIDs(ctx, cursor, userIDs, int64(req.Count))
		if err != nil {
			return nil, err
		}

	case constant.Subscribe:
		userIDs, err := o.Database.GetSubscriberUserIDs(ctx, userID)
		if err != nil {
			return nil, err
		}
		postsDB, nextCursor, err = o.Database.GetPostsByCursorAndUserIDs(ctx, cursor, userIDs, int64(req.Count))
		if err != nil {
			return nil, err
		}
	case constant.Reply:
		postIds, err := o.Database.GetCommentPostIDsByUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		postsDB, nextCursor, err = o.Database.GetPostsByCursorAndPostIDs(ctx, cursor, postIds, int64(req.Count))
		if err != nil {
			return nil, err
		}
	case constant.Like:
		postIds, err := o.Database.GetPostIDsByLike(ctx, userID)
		if err != nil {
			return nil, err
		}
		postsDB, nextCursor, err = o.Database.GetPostsByCursorAndPostIDs(ctx, cursor, postIds, int64(req.Count))
		if err != nil {
			return nil, err
		}
	case constant.Collect:
		postIds, err := o.Database.GetPostIDsByCollect(ctx, userID)
		if err != nil {
			return nil, err
		}
		postsDB, nextCursor, err = o.Database.GetPostsByCursorAndPostIDs(ctx, cursor, postIds, int64(req.Count))
		if err != nil {
			return nil, err
		}
	default:
		postsDB, nextCursor, err = o.Database.GetPostsByCursorAndUser(ctx, cursor, userID, int64(req.Count))
		if err != nil {
			return nil, err
		}
	}
	postsPB := convert.PostsDB2Pb(postsDB)

	resp.Posts = postsPB
	if nextCursor != "" {
		nextCursorInt, err := strconv.ParseInt(nextCursor, 10, 64)
		if err != nil {
			return nil, err
		}
		resp.NextCursor = nextCursorInt
	}
	return resp, nil
}

func (o *chatSvr) GetCommentPostListByPostID(ctx context.Context, req *chatpb.GetCommentPostListByPostIDReq) (*chatpb.GetCommentPostListByPostIDResp, error) {
	resp := &chatpb.GetCommentPostListByPostIDResp{}
	postsDB, nextCursor, err := o.Database.GetCommentPostsByPostID(ctx, req.NextCursor, req.PostID, int64(req.Count))
	if err != nil {
		return nil, err
	}
	postsPB := convert.PostsDB2Pb(postsDB)

	resp.Posts = postsPB
	if nextCursor != "" {
		nextCursorInt, err := strconv.ParseInt(nextCursor, 10, 64)
		if err != nil {
			return nil, err
		}
		resp.NextCursor = nextCursorInt
	}
	return resp, nil
}

func (o *chatSvr) GenPostID(ctx context.Context, postID *string) error {
	if *postID != "" {
		_, err := o.Database.GetPostByID(ctx, *postID)
		if err == nil {
			return servererrs.ErrGroupIDExisted.WrapMsg("post id existed " + *postID)
		} else if IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}
	for i := 0; i < 10; i++ {
		id := encrypt.Md5(strings.Join([]string{mcontext.GetOperationID(ctx), strconv.FormatInt(time.Now().UnixNano(), 10), strconv.Itoa(rand.Int())}, ",;,"))
		bi := big.NewInt(0)
		bi.SetString(id[0:8], 16)
		id = bi.String()
		_, err := o.Database.GetPostByID(ctx, id)
		if err == nil {
			continue
		} else if IsNotFound(err) {
			*postID = id
			return nil
		} else {
			return err
		}
	}
	return servererrs.ErrData.WrapMsg("group id gen error")
}
