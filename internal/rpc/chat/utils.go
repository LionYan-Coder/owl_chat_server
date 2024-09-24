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
	"crypto/rand"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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

func DbToPbAttributes(attributes []*chat.Attribute) []*common.UserPublicInfo {
	return datautil.Slice(attributes, DbToPbAttribute)
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

func generateNonce(size int) (string, error) {
	nonce := make([]byte, size)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return hex.EncodeToString(nonce), nil
}

func validateSignature(publicKey, nonce, signature string) (b bool, err error) {
	pubKey, err := hexutil.Decode(publicKey)
	if err != nil {
		return false, err
	}
	message, err := hexutil.Decode(nonce)
	if err != nil {
		return false, err
	}
	sig, err := hexutil.Decode(signature)
	if err != nil {
		return false, err
	}
	b = crypto.VerifySignature(pubKey, message, sig)
	return b, nil
}
