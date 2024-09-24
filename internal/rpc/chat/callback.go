// Copyright © 2023 OpenIM open source community. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openimsdk/chat/pkg/common/constant"
	"github.com/openimsdk/chat/pkg/eerrs"
	"github.com/openimsdk/chat/pkg/protocol/chat"
	constantpb "github.com/openimsdk/chat/pkg/protocol/constant"
	"github.com/openimsdk/chat/pkg/redpacket"
	"github.com/openimsdk/chat/pkg/redpacket/servererrs"
	"github.com/openimsdk/tools/errs"
)

type CallbackBeforeAddFriendReq struct {
	CallbackCommand `json:"callbackCommand"`
	FromUserID      string `json:"fromUserID" `
	ToUserID        string `json:"toUserID"`
	ReqMsg          string `json:"reqMsg"`
	Remark          string `json:"remark"`
	OperationID     string `json:"operationID"`
}

type CallbackCommand string

func (c CallbackCommand) GetCallbackCommand() string {
	return string(c)
}

func (o *chatSvr) OpenIMCallback(ctx context.Context, req *chat.OpenIMCallbackReq) (*chat.OpenIMCallbackResp, error) {
	var result *chat.OpenIMCallbackResp
	var err error

	switch req.Command {
	case constantpb.CallbackBeforeAddFriendCommand:
		result, err = o.handleCallbackBeforeAddFriend(ctx, req)
	case constantpb.CallbackBeforeSendSingleMsgCommand:
		result, err = o.handleCallbackBeforeMsg(ctx, req)
	case constantpb.CallbackBeforeSendGroupMsgCommand:
		result, err = o.handleCallbackBeforeMsg(ctx, req)
	default:
		return nil, errs.ErrArgs.WrapMsg(fmt.Sprintf("invalid command %s", req.Command))
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (o *chatSvr) handleCallbackBeforeMsg(ctx context.Context, req *chat.OpenIMCallbackReq) (*chat.OpenIMCallbackResp, error) {
	var data constantpb.CommonCallbackReq
	if err := json.Unmarshal([]byte(req.Body), &data); err != nil {
		return nil, errs.Wrap(err)
	}
	if data.MsgFrom == constantpb.UserMsgType && data.ContentType == constantpb.Custom {
		var content map[string]interface{}
		if err := json.Unmarshal([]byte(data.Content), &content); err != nil {
			return nil, errs.Wrap(err)
		}
		var data1 map[string]interface{}
		if err := json.Unmarshal([]byte(content["data"].(string)), &data1); err != nil {
			return nil, errs.Wrap(err)
		}
		if customType, ok := data1["customType"].(float64); ok {
			intCustomType := int32(customType)
			if intCustomType >= constantpb.SendPrivateRedPacket {
				if intCustomType <= constantpb.SendExclusiveRedPacket {
					return o.SendRedPacket(ctx, &data)
				}
				if intCustomType <= constantpb.ReceiveExclusiveRedPacket {
					return o.ReceiveRedPacket(ctx, &data)
				}
			}
		}
	}
	return &chat.OpenIMCallbackResp{
		ActionCode: 0,
		NextCode:   0,
	}, nil
}

func (o *chatSvr) handleCallbackBeforeAddFriend(ctx context.Context, req *chat.OpenIMCallbackReq) (*chat.OpenIMCallbackResp, error) {
	var data CallbackBeforeAddFriendReq
	if err := json.Unmarshal([]byte(req.Body), &data); err != nil {
		return nil, errs.Wrap(err)
	}
	user, err := o.Database.GetAttribute(ctx, data.ToUserID)
	if err != nil {
		return nil, err
	}
	if user.AllowAddFriend != constant.OrdinaryUserAddFriendEnable {
		return nil, eerrs.ErrRefuseFriend.WrapMsg(fmt.Sprintf("state %d", user.AllowAddFriend))
	}
	return &chat.OpenIMCallbackResp{
		ActionCode: 0,
		NextCode:   0,
	}, nil
}

// 发送红包
func (o *chatSvr) SendRedPacket(ctx context.Context, msgData *constantpb.CommonCallbackReq) (*chat.OpenIMCallbackResp, error) {
	var content map[string]interface{}
	if err := json.Unmarshal([]byte(msgData.Content), &content); err != nil {
		return nil, errs.Wrap(err)
	}

	var contentData map[string]interface{}
	if err := json.Unmarshal([]byte(content["data"].(string)), &contentData); err != nil {
		return nil, errs.Wrap(err)
	}

	if params, ok := contentData["data"].(map[string]interface{}); ok {
		redpacketResp := &redpacket.Response{}
		params["clientMsgID"] = msgData.ClientMsgID
		if err := o.RedPacketClient.SyncPost(ctx, msgData.SendID, "/redPacket/send", params, redpacketResp, &o.Share.RedPacket); err != nil {
			return &chat.OpenIMCallbackResp{
				ActionCode: 0,
				NextCode:   1,
				ErrDlt:     err.Error(),
			}, nil
		}

		if redpacketResp.Code == redpacket.SuccessCode {
			return &chat.OpenIMCallbackResp{
				ActionCode: 0,
				NextCode:   0,
			}, nil
		} else {
			return &chat.OpenIMCallbackResp{
				ActionCode: 0,
				NextCode:   1,
				ErrDlt:     redpacketResp.Msg,
				ErrMsg:     redpacketResp.Msg,
				ErrCode:    servererrs.ServerInternalError,
			}, nil
		}
	}

	return &chat.OpenIMCallbackResp{
		ActionCode: 0,
		NextCode:   0,
	}, nil
}

// 领取红包
func (o *chatSvr) ReceiveRedPacket(ctx context.Context, msgData *constantpb.CommonCallbackReq) (*chat.OpenIMCallbackResp, error) {
	var content map[string]interface{}
	if err := json.Unmarshal([]byte(msgData.Content), &content); err != nil {
		return nil, errs.Wrap(err)
	}

	resp := &chat.OpenIMCallbackResp{
		ActionCode: 0,
		NextCode:   0,
	}

	if contentData, ok := content["data"].(string); ok {
		var innerData map[string]interface{}
		if err := json.Unmarshal([]byte(contentData), &innerData); err != nil {
			return nil, errs.Wrap(err)
		}
		if data, ok := innerData["data"].(map[string]interface{}); ok {
			if redPacketId, ok := data["redPacketId"].(string); ok {
				// 获取领取红包参数
				receiveParams := map[string]interface{}{"redPacketId": redPacketId}
				redpacketResp := &redpacket.Response{}
				if err := o.RedPacketClient.SyncPost(ctx, msgData.SendID, "/redPacket/receive", receiveParams, redpacketResp, &o.Share.RedPacket); err != nil {
					return &chat.OpenIMCallbackResp{
						ActionCode: 1,
						ErrDlt:     err.Error(),
					}, nil
				}

				if redpacketResp.Code == redpacket.SuccessCode {
					return &chat.OpenIMCallbackResp{
						ActionCode: 0,
						NextCode:   0,
					}, nil
				} else {
					return &chat.OpenIMCallbackResp{
						ActionCode: 0,
						NextCode:   1,
						ErrDlt:     redpacketResp.Msg,
						ErrMsg:     redpacketResp.Msg,
						ErrCode:    servererrs.ServerInternalError,
					}, nil
				}
			}
		}
	}

	return resp, nil
}
