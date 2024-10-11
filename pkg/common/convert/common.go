package convert

import (
	"github.com/openimsdk/chat/pkg/common/db/table/chat"
	"github.com/openimsdk/chat/pkg/protocol/common"
)

func PictureElemPb2DB(postPicturePB *common.PictureElem) *chat.PostPicture {
	return &chat.PostPicture{
		SourcePath:      postPicturePB.SourcePath,
		SourcePicture:   *PictureBaseInfoPb2DB(postPicturePB.SourcePicture),
		BigPicture:      *PictureBaseInfoPb2DB(postPicturePB.BigPicture),
		SnapshotPicture: *PictureBaseInfoPb2DB(postPicturePB.SnapshotPicture),
	}
}

func PictureElemDB2Pb(postPictureDB *chat.PostPicture) *common.PictureElem {
	return &common.PictureElem{
		SourcePath:      postPictureDB.SourcePath,
		SourcePicture:   PictureBaseInfoDB2Pb(&postPictureDB.SourcePicture),
		BigPicture:      PictureBaseInfoDB2Pb(&postPictureDB.BigPicture),
		SnapshotPicture: PictureBaseInfoDB2Pb(&postPictureDB.SnapshotPicture),
	}
}

func VideoElemPb2DB(postVideoPB *common.VideoElem) *chat.PostVideo {
	return &chat.PostVideo{
		VideoPath:      postVideoPB.VideoPath,
		VideoUUID:      postVideoPB.VideoUUID,
		VideoURL:       postVideoPB.VideoUrl,
		VideoType:      postVideoPB.VideoType,
		VideoSize:      postVideoPB.VideoSize,
		Duration:       postVideoPB.Duration,
		SnapshotPath:   postVideoPB.SnapshotPath,
		SnapshotUUID:   postVideoPB.SnapshotUUID,
		SnapshotSize:   postVideoPB.SnapshotSize,
		SnapshotURL:    postVideoPB.SnapshotUrl,
		SnapshotWidth:  postVideoPB.SnapshotWidth,
		SnapshotHeight: postVideoPB.SnapshotHeight,
		SnapshotType:   postVideoPB.SnapshotType,
	}
}

func VideoElemDB2Pb(postVideoDB *chat.PostVideo) *common.VideoElem {
	return &common.VideoElem{
		VideoPath:      postVideoDB.VideoPath,
		VideoUUID:      postVideoDB.VideoUUID,
		VideoUrl:       postVideoDB.VideoURL,
		VideoType:      postVideoDB.VideoType,
		VideoSize:      postVideoDB.VideoSize,
		Duration:       postVideoDB.Duration,
		SnapshotPath:   postVideoDB.SnapshotPath,
		SnapshotUUID:   postVideoDB.SnapshotUUID,
		SnapshotSize:   postVideoDB.SnapshotSize,
		SnapshotUrl:    postVideoDB.SnapshotURL,
		SnapshotWidth:  postVideoDB.SnapshotWidth,
		SnapshotHeight: postVideoDB.SnapshotHeight,
		SnapshotType:   postVideoDB.SnapshotType,
	}
}

func PictureBaseInfoDB2Pb(postPictureBaseInfoDB *chat.PictureBaseInfo) *common.PictureBaseInfo {
	return &common.PictureBaseInfo{
		Uuid:   postPictureBaseInfoDB.UUID,
		Type:   postPictureBaseInfoDB.Type,
		Size:   postPictureBaseInfoDB.Size,
		Width:  postPictureBaseInfoDB.Width,
		Height: postPictureBaseInfoDB.Height,
		Url:    postPictureBaseInfoDB.URL,
	}
}

func PictureBaseInfoPb2DB(postPictureBaseInfoPB *common.PictureBaseInfo) *chat.PictureBaseInfo {
	return &chat.PictureBaseInfo{
		UUID:   postPictureBaseInfoPB.Uuid,
		Type:   postPictureBaseInfoPB.Type,
		Size:   postPictureBaseInfoPB.Size,
		Width:  postPictureBaseInfoPB.Width,
		Height: postPictureBaseInfoPB.Height,
		URL:    postPictureBaseInfoPB.Url,
	}
}
