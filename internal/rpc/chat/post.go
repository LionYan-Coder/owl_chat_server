package chat

import (
	"context"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/openimsdk/chat/pkg/redpacket/servererrs"
	"github.com/openimsdk/tools/mcontext"
	"github.com/openimsdk/tools/utils/encrypt"

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
		UserID:        userID,
		ForwardPostID: req.ForwardPostID,
		AllowComment:  req.AllowComment,
		AllowForward:  req.AllowForward,
		Content:       req.Content.Value,
		AtUserIds:     req.AtUserIds,
		MediaMsgs:     convert.PostMediasPb2DB(req.MediaMsgs),
	}
	if err := o.GenPostID(ctx, &postDB.PostID); err != nil {
		return nil, err
	}
	err = o.Database.CreatePost(ctx, []*chat.PostDB{postDB})
	if err != nil {
		return nil, err
	}
	return &chatpb.PublishPostResp{}, nil
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

	relation, err := o.Database.GetUserPostRelation(ctx, opUserID, post.PostID)
	if err != nil {
		return nil, err
	}
	if relation == nil {
		relation = &chat.UserPostRelation{
			UserID:  opUserID,
			PostID:  post.PostID,
			IsLiked: constant.Liked,
		}
		err = o.Database.CreateUserPostRelation(ctx, []*chat.UserPostRelation{relation})
		if err != nil {
			return nil, err
		}
	} else {
		if relation.IsLiked == constant.Liked {
			relation.IsLiked = constant.NotLiked
		} else {
			relation.IsLiked = constant.Liked
		}
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

	relation, err := o.Database.GetUserPostRelation(ctx, opUserID, post.PostID)
	if err != nil {
		return nil, err
	}
	if relation == nil {
		relation = &chat.UserPostRelation{
			UserID:      opUserID,
			PostID:      post.PostID,
			IsCollected: constant.Collected,
		}
		err = o.Database.CreateUserPostRelation(ctx, []*chat.UserPostRelation{relation})
		if err != nil {
			return nil, err
		}
	} else {
		if relation.IsCollected == constant.Collected {
			relation.IsCollected = constant.NotCollected
		} else {
			relation.IsCollected = constant.Collected
		}
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

	err = o.Database.DeletePost(ctx, req.PostID)
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

func (o *chatSvr) GetPostPaginationByUser(ctx context.Context, req *chatpb.GetPostPaginationByUserReq) (*chatpb.GetPostPaginationByUserResp, error) {
	resp := &chatpb.GetPostPaginationByUserResp{}
	count, posts, err := o.Database.GetPostPaginationByUser(ctx, req.UserID, req.Pagination)
	if err != nil {
		return nil, err
	}
	postsPB := convert.PostsDB2Pb(posts)

	resp.Posts = postsPB
	resp.Total = count
	return resp, nil
}

func (o *chatSvr) GetPostPagination(ctx context.Context, req *chatpb.GetPostPaginationReq) (*chatpb.GetPostPaginationResp, error) {
	// userID, err := mctx.CheckUser(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	resp := &chatpb.GetPostPaginationResp{}

	var total int64
	var postsDB []*chat.Post
	var err error
	total, postsDB, err = o.Database.GetPostPagination(ctx, req.Pagination)

	if err != nil {
		return nil, err
	}
	postsPB := convert.PostsDB2Pb(postsDB)

	resp.Posts = postsPB
	resp.Total = total
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
