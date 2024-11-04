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

package imapi

import (
	"context"
	"sync"
	"time"

	"github.com/openimsdk/chat/pkg/eerrs"
	"github.com/openimsdk/chat/pkg/protocol/auth"
	"github.com/openimsdk/chat/pkg/protocol/friend"
	"github.com/openimsdk/chat/pkg/protocol/group"
	"github.com/openimsdk/chat/pkg/protocol/sdkwss"
	"github.com/openimsdk/chat/pkg/protocol/user"
	"github.com/openimsdk/protocol/msggateway"
	"github.com/openimsdk/tools/log"

	chatpb "github.com/openimsdk/chat/pkg/protocol/chat"
	constantpb "github.com/openimsdk/chat/pkg/protocol/constant"
)

type CallerInterface interface {
	ImAdminTokenWithDefaultAdmin(ctx context.Context) (string, error)
	ImportFriend(ctx context.Context, ownerUserID string, friendUserID []string) error
	UserToken(ctx context.Context, userID string, platform int32) (string, error)
	InviteToGroup(ctx context.Context, userID string, groupIDs []string) error
	UpdateUserInfo(ctx context.Context, userID string, nickName string, faceURL string, coverURL string, about string, account string) error
	ForceOffLine(ctx context.Context, userID string) error
	RegisterUser(ctx context.Context, users []*sdkwss.UserInfo) error
	FindGroupInfo(ctx context.Context, groupIDs []string) ([]*sdkwss.GroupInfo, error)
	UserRegisterCount(ctx context.Context, start int64, end int64) (map[string]int64, int64, error)
	FriendUserIDs(ctx context.Context, userID string) ([]string, error)
	AccountCheckSingle(ctx context.Context, userID string) (bool, error)
	UserOlineStatus(ctx context.Context, userIDs []string) ([]msggateway.GetUsersOnlineStatusResp_SuccessResult, error)
	UserOlineTimes(ctx context.Context, userIDs []string) (*chatpb.GetUsersTimeResp, error)
}

type Caller struct {
	imApi           string
	imSecret        string
	defaultIMUserID string
	token           string
	timeout         time.Time
	lock            sync.Mutex
}

func New(imApi string, imSecret string, defaultIMUserID string) CallerInterface {
	return &Caller{
		imApi:           imApi,
		imSecret:        imSecret,
		defaultIMUserID: defaultIMUserID,
	}
}

func (c *Caller) ImportFriend(ctx context.Context, ownerUserID string, friendUserIDs []string) error {
	if len(friendUserIDs) == 0 {
		return nil
	}
	_, err := importFriend.Call(ctx, c.imApi, &friend.ImportFriendReq{
		OwnerUserID:   ownerUserID,
		FriendUserIDs: friendUserIDs,
	})
	return err
}

func (c *Caller) ImAdminTokenWithDefaultAdmin(ctx context.Context) (string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.token == "" || c.timeout.Before(time.Now()) {
		userID := c.defaultIMUserID
		token, err := c.UserToken(ctx, userID, constantpb.AdminPlatformID)
		if err != nil {
			log.ZError(ctx, "get im admin token", err, "userID", userID)
			return "", err
		}
		log.ZDebug(ctx, "get im admin token", "userID", userID)
		c.token = token
		c.timeout = time.Now().Add(time.Minute * 5)
	}
	return c.token, nil
}

func (c *Caller) UserToken(ctx context.Context, userID string, platformID int32) (string, error) {
	resp, err := userToken.Call(ctx, c.imApi, &auth.UserTokenReq{
		Secret:     c.imSecret,
		PlatformID: platformID,
		UserID:     userID,
	})
	if err != nil {
		return "", err
	}
	return resp.Token, nil
}

func (c *Caller) ParseToken(ctx context.Context, token string) (*auth.ParseTokenResp, error) {
	resp, err := parseToken.Call(ctx, c.imApi, &auth.ParseTokenReq{
		Token: token,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Caller) InviteToGroup(ctx context.Context, userID string, groupIDs []string) error {
	for _, groupID := range groupIDs {
		_, _ = inviteToGroup.Call(ctx, c.imApi, &group.InviteUserToGroupReq{
			GroupID:        groupID,
			Reason:         "",
			InvitedUserIDs: []string{userID},
		})
	}
	return nil
}

func (c *Caller) UpdateUserInfo(ctx context.Context, userID string, nickName string, faceURL string, coverURL string, about string, account string) error {
	_, err := updateUserInfo.Call(ctx, c.imApi, &user.UpdateUserInfoReq{UserInfo: &sdkwss.UserInfo{
		UserID:   userID,
		Nickname: nickName,
		FaceURL:  faceURL,
		CoverURL: coverURL,
		About:    about,
		Account:  account,
	}})
	return err
}

func (c *Caller) RegisterUser(ctx context.Context, users []*sdkwss.UserInfo) error {
	_, err := registerUser.Call(ctx, c.imApi, &user.UserRegisterReq{
		Secret: c.imSecret,
		Users:  users,
	})
	return err
}

func (c *Caller) ForceOffLine(ctx context.Context, userID string) error {
	for id := range constantpb.PlatformID2Name {
		_, _ = forceOffLine.Call(ctx, c.imApi, &auth.ForceLogoutReq{
			PlatformID: int32(id),
			UserID:     userID,
		})
	}
	return nil
}

func (c *Caller) FindGroupInfo(ctx context.Context, groupIDs []string) ([]*sdkwss.GroupInfo, error) {
	resp, err := getGroupsInfo.Call(ctx, c.imApi, &group.GetGroupsInfoReq{
		GroupIDs: groupIDs,
	})
	if err != nil {
		return nil, err
	}
	return resp.GroupInfos, nil
}

func (c *Caller) UserRegisterCount(ctx context.Context, start int64, end int64) (map[string]int64, int64, error) {
	resp, err := registerUserCount.Call(ctx, c.imApi, &user.UserRegisterCountReq{
		Start: start,
		End:   end,
	})
	if err != nil {
		return nil, 0, err
	}
	return resp.Count, resp.Total, nil
}

func (c *Caller) FriendUserIDs(ctx context.Context, userID string) ([]string, error) {
	resp, err := friendUserIDs.Call(ctx, c.imApi, &friend.GetFriendIDsReq{UserID: userID})
	if err != nil {
		return nil, err
	}
	return resp.FriendIDs, nil
}

// return true when isUserNotExist.
func (c *Caller) AccountCheckSingle(ctx context.Context, userID string) (bool, error) {
	resp, err := accountCheck.Call(ctx, c.imApi, &user.AccountCheckReq{CheckUserIDs: []string{userID}})
	if err != nil {
		return false, err
	}
	if resp.Results[0].AccountStatus == "registered" {
		return false, eerrs.ErrAccountAlreadyRegister.Wrap()
	}
	return true, nil
}

func (c *Caller) UserOlineStatus(ctx context.Context, userIDs []string) ([]msggateway.GetUsersOnlineStatusResp_SuccessResult, error) {
	resp, err := allUserOnlineStatus.Call(ctx, c.imApi, &msggateway.GetUsersOnlineStatusReq{UserIDs: userIDs})
	if err != nil {
		return nil, err
	}
	return *resp, nil
}

func (c *Caller) UserOlineTimes(ctx context.Context, userIDs []string) (*chatpb.GetUsersTimeResp, error) {
	resp, err := usersOnlineTime.Call(ctx, c.imApi, &chatpb.GetUsersTimeReq{UserIDs: userIDs})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
