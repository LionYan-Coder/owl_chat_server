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

package redpacket

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/openimsdk/chat/pkg/common/config"
	"github.com/openimsdk/chat/pkg/redpacket/servererrs"
	"github.com/openimsdk/protocol/constant"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/mq/memamq"
	"github.com/openimsdk/tools/utils/httputil"
)

type Client struct {
	client *httputil.HTTPClient
	url    string
	queue  *memamq.MemoryQueue
}

type Response struct {
	Msg  string      `json:"msg"`  // 消息内容
	Code int         `json:"code"` // 状态码
	Data interface{} `json:"data"` // 返回的数据
}

const (
	SuccessCode = 200
)

const (
	webhookWorkerCount = 2
	webhookBufferSize  = 100
)

func NewRedPacketClient(url string, options ...*memamq.MemoryQueue) *Client {
	var queue *memamq.MemoryQueue
	if len(options) > 0 && options[0] != nil {
		queue = options[0]
	} else {
		queue = memamq.NewMemoryQueue(webhookWorkerCount, webhookBufferSize)
	}

	http.DefaultTransport.(*http.Transport).MaxConnsPerHost = 100 // Enhance the default number of max connections per host

	return &Client{
		client: httputil.NewHTTPClient(httputil.NewClientConfig()),
		url:    url,
		queue:  queue,
	}
}

func (c *Client) SyncPost(ctx context.Context, userID string, url string, req interface{}, resp *Response, config *config.RpcRedPacket) error {
	return c.post(ctx, userID, url, req, resp, config.Timeout)
}

func (c *Client) post(ctx context.Context, userID string, url string, input interface{}, output *Response, timeout int) error {
	fullURL := c.url + url
	log.ZInfo(ctx, "redpacket", "url", fullURL, "input", input, "config", timeout)
	b, err := c.client.Post(ctx, fullURL, map[string]string{constant.OperationID: userID}, input, timeout)
	if err != nil {
		log.ZInfo(ctx, "webhook redpacket error", servererrs.ErrNetwork.WrapMsg(err.Error(), "post url", fullURL))
		return servererrs.ErrNetwork.WrapMsg(err.Error(), "post url", fullURL)
	}
	if err = json.Unmarshal(b, output); err != nil {
		log.ZInfo(ctx, "webhook redpacket json error", servererrs.ErrData.WithDetail(err.Error()+" response format error"))
		return servererrs.ErrData.WithDetail(err.Error() + " response format error")
	}
	log.ZInfo(ctx, "webhook redpacket success", "url", fullURL, "input", input, "response", string(b))
	return nil
}
