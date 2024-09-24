package chat

import (
	"context"

	chatdb "github.com/openimsdk/chat/pkg/common/db/table/chat"
	"github.com/openimsdk/chat/pkg/common/mctx"
	"github.com/openimsdk/chat/pkg/protocol/chat"
	"github.com/openimsdk/tools/errs"
	"go.mongodb.org/mongo-driver/mongo"
)

func (o *chatSvr) GetGroupFromContact(ctx context.Context, req *chat.GetGroupFromContactReq) (*chat.GetGroupFromContactResp, error) {
	userID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}
	contact := &chatdb.Contact{}
	contact, err = o.Database.GetGroupFromContact(ctx, userID)
	if err != nil && errs.Unwrap(err) == mongo.ErrNoDocuments {
		return &chat.GetGroupFromContactResp{
			GroupIDs: []string{},
		}, nil
	}
	if err != nil {
		return nil, err
	}

	if contact.GroupIDs == nil {
		contact.GroupIDs = []string{}
	}

	return &chat.GetGroupFromContactResp{GroupIDs: contact.GroupIDs}, nil
}

func (o *chatSvr) DeleteGroupFromContact(ctx context.Context, req *chat.DeleteGroupFromContactReq) (resp *chat.DeleteGroupFromContactResp, err error) {
	userID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}
	err = o.Database.DeleteGroupFromContact(ctx, userID, req.GroupIDs)

	return nil, err
}

func (o *chatSvr) SaveGroupToContact(ctx context.Context, req *chat.SaveGroupToContactReq) (resp *chat.SaveGroupToContactResp, err error) {
	userID, err := mctx.CheckUser(ctx)
	if err != nil {
		return nil, err
	}
	err = o.Database.SaveGroupToContact(ctx, userID, req.GroupIDs)
	return nil, err
}

// DeleteUserGroupApplicationFromAll implements chat.ChatServer.
func (o *chatSvr) DeleteUserGroupApplicationFromAll(context.Context, *chat.DeleteGroupApplicationFromAlltReq) (*chat.DeleteGroupApplicationFromAllResp, error) {
	return nil, nil
}

// DeleteUserGroupApplicationFromApplicant implements chat.ChatServer.
func (o *chatSvr) DeleteUserGroupApplicationFromApplicant(context.Context, *chat.DeleteGroupApplicationFromApplicantReq) (*chat.DeleteGroupApplicationFromApplicantResp, error) {
	return nil, nil
}

// DeleteUserGroupApplicationFromRecipient implements chat.ChatServer.
func (o *chatSvr) DeleteUserGroupApplicationFromRecipient(context.Context, *chat.DeleteGroupApplicationFromRecipientReq) (*chat.DeleteGroupApplicationFromRecipientResp, error) {
	return nil, nil
}
