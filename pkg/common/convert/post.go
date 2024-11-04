package convert

import (
	"github.com/openimsdk/chat/pkg/common/db/table/chat"
	chatpb "github.com/openimsdk/chat/pkg/protocol/chat"
	"github.com/openimsdk/chat/pkg/protocol/common"
	"github.com/openimsdk/tools/utils/datautil"

	"github.com/openimsdk/chat/pkg/common/constant"
)

func PostDB2Pb(postDB *chat.Post) *chatpb.Post {
	if postDB == nil {
		return nil
	}
	postPB := &chatpb.Post{}
	if err := datautil.CopyStructFields(postPB, postDB); err != nil {
		return nil
	}
	if postDB.ForwardPost != nil {
		postPB.ForwardPost = PostDB2Pb(postDB.ForwardPost)
	}
	if postDB.CommentPost != nil {
		postPB.CommentPost = PostDB2Pb(postDB.CommentPost)
	}
	if postDB.RefPost != nil {
		postPB.RefPost = PostDB2Pb(postDB.RefPost)
	}
	postPB.CreateTime = postDB.CreateTime.UnixMilli()
	postPB.UpdateTime = postDB.UpdateTime.UnixMilli()
	postPB.UserInfo = DbToPbAttribute(postDB.UserInfo)
	postPB.AtUserInfoList = DbToPbAttributes(postDB.AtUserInfoList)
	postPB.MediaMsgs = PostMediasDB2Pb(postDB.MediaMsgs)
	return postPB
}

func PostsDB2Pb(postsDB []*chat.Post) []*chatpb.Post {
	return datautil.Slice(postsDB, PostDB2Pb)
}

func PostPb2DB(postPB *chatpb.Post) *chat.Post {
	postDB := &chat.Post{}
	if err := datautil.CopyStructFields(postDB, postPB); err != nil {
		return nil
	}
	postDB.MediaMsgs = PostMediasPb2DB(postPB.MediaMsgs)
	return postDB
}

func PostsPb2DB(postPB []*chatpb.Post) []*chat.Post {
	return datautil.Slice(postPB, PostPb2DB)
}

func PostMediaDB2Pb(mediaDB *chat.PostMedia) *common.PostMedia {
	return &common.PostMedia{
		MediaType:   mediaDB.MediaType,
		PostPicture: PictureElemDB2Pb(&mediaDB.PostPicture),
		PostVideo:   VideoElemDB2Pb(&mediaDB.PostVideo),
	}
}

func PostMediasDB2Pb(mediasDB []*chat.PostMedia) (mediasPB []*common.PostMedia) {
	for _, mediaDB := range mediasDB {
		mediasPB = append(mediasPB, PostMediaDB2Pb(mediaDB))
	}
	return mediasPB
}

func PostMediasPb2DB(mediasPB []*common.PostMedia) (mediasDB []*chat.PostMedia) {
	for _, mediaPB := range mediasPB {
		mediasDB = append(mediasDB, PostMediaPb2DB(mediaPB))
	}
	return mediasDB
}

func PostMediaPb2DB(mediaPB *common.PostMedia) *chat.PostMedia {
	media := &chat.PostMedia{MediaType: mediaPB.MediaType}

	switch mediaPB.MediaType {
	case constant.PostMediaTypePicture:
		media.PostPicture = *PictureElemPb2DB(mediaPB.PostPicture)
	case constant.PostMediaTypeVideo:
		media.PostVideo = *VideoElemPb2DB(mediaPB.PostVideo)
	}

	return media
}
