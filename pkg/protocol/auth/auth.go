// Copyright © 2023 OpenIM. All rights reserved.
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

package auth

import (
	"errors"
	"github.com/openimsdk/chat/pkg/protocol/constant"
)

func (x *UserTokenReq) Check() error {
	if x.UserID == "" {
		return errors.New("userID is empty")
	}
	if x.PlatformID > constant.AdminPlatformID || x.PlatformID < constant.IOSPlatformID {
		return errors.New("platform is invalidate")
	}
	return nil
}

func (x *ForceLogoutReq) Check() error {
	if x.UserID == "" {
		return errors.New("userID is empty")
	}
	if x.PlatformID > constant.AdminPlatformID || x.PlatformID < constant.IOSPlatformID {
		return errors.New("platformID is invalidate")
	}
	return nil
}

func (x *ParseTokenReq) Check() error {
	if x.Token == "" {
		return errors.New("userID is empty")
	}
	return nil
}

func (x *GetUserTokenReq) Check() error {
	if x.UserID == "" {
		errors.New("userID is empty")
	}

	if x.PlatformID == 0 {
		errors.New("platformID is empty")
	}
	return nil
}
