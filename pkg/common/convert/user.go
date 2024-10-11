package convert

import (
	"github.com/openimsdk/chat/pkg/common/db/table/chat"
	"github.com/openimsdk/chat/pkg/protocol/common"
	"github.com/openimsdk/tools/utils/datautil"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func DbToPbAttribute(attribute *chat.Attribute) *common.UserPublicInfo {
	if attribute == nil {
		return nil
	}
	return &common.UserPublicInfo{
		UserID:    attribute.UserID,
		Account:   attribute.Account,
		Address:   attribute.Address,
		Nickname:  attribute.Nickname,
		FaceURL:   attribute.FaceURL,
		CoverURL:  attribute.CoverURL,
		About:     attribute.About,
		PublicKey: attribute.PublicKey,
	}
}

func PbToDbAttribute(attribute *common.UserPublicInfo) *chat.Attribute {
	if attribute == nil {
		return nil
	}
	return &chat.Attribute{
		UserID:    attribute.UserID,
		Account:   attribute.Account,
		Address:   attribute.Address,
		Nickname:  attribute.Nickname,
		FaceURL:   attribute.FaceURL,
		CoverURL:  attribute.CoverURL,
		About:     attribute.About,
		PublicKey: attribute.PublicKey,
	}
}

func DbToPbAttributes(attributes []*chat.Attribute) []*common.UserPublicInfo {
	return datautil.Slice(attributes, DbToPbAttribute)
}

func PbToDbAttributes(attributes []*common.UserPublicInfo) []*chat.Attribute {
	return datautil.Slice(attributes, PbToDbAttribute)
}

func DbToPbUserFullInfo(attribute *chat.Attribute) *common.UserFullInfo {
	createTimeProto := timestamppb.New(attribute.CreateTime)
	return &common.UserFullInfo{
		UserID: attribute.UserID,
		// Password:         "",
		Account:          attribute.Account,
		Address:          attribute.Address,
		Nickname:         attribute.Nickname,
		FaceURL:          attribute.FaceURL,
		CoverURL:         attribute.CoverURL,
		About:            attribute.About,
		CreateTime:       createTimeProto,
		PublicKey:        attribute.PublicKey,
		AllowAddFriend:   attribute.AllowAddFriend,
		AllowBeep:        attribute.AllowBeep,
		AllowVibration:   attribute.AllowVibration,
		GlobalRecvMsgOpt: attribute.GlobalRecvMsgOpt,
		RegisterType:     attribute.RegisterType,
	}
}

func DbToPbUserFullInfos(attributes []*chat.Attribute) []*common.UserFullInfo {
	return datautil.Slice(attributes, DbToPbUserFullInfo)
}
