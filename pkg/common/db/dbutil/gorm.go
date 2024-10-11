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

package dbutil

import (
	"context"

	"github.com/openimsdk/tools/db/mongoutil"
	"github.com/openimsdk/tools/db/pagination"
	"github.com/openimsdk/tools/errs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func IsDBNotFound(err error) bool {
	return errs.Unwrap(err) == mongo.ErrNoDocuments
}

func FindPageWithAggregation[T any](ctx context.Context, coll *mongo.Collection, filter any, pagination pagination.Pagination, pipeline mongo.Pipeline, opts ...*options.AggregateOptions) (int64, []T, error) {
	count, err := mongoutil.Count(ctx, coll, filter, options.Count())
	if err != nil {
		return 0, nil, errs.WrapMsg(err, "mongo failed to count documents in collection")
	}

	// 如果没有文档或者分页参数无效，直接返回
	if count == 0 || pagination == nil {
		return count, nil, nil
	}

	// 计算分页的 skip 和 limit
	skip := int64(pagination.GetPageNumber()-1) * int64(pagination.GetShowNumber())
	if skip < 0 || skip >= count || pagination.GetShowNumber() <= 0 {
		return count, nil, nil
	}

	pipeline = append(pipeline,
		bson.D{{Key: "$skip", Value: skip}},                        // 跳过指定数量的文档
		bson.D{{Key: "$limit", Value: pagination.GetShowNumber()}}, // 限制返回文档的数量
	)

	// 执行聚合查询
	cursor, err := coll.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return 0, nil, err
	}
	defer cursor.Close(ctx)

	// 解析聚合结果
	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return 0, nil, err
	}

	return count, results, nil
}
