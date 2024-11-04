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
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/openimsdk/tools/errs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func IsDBNotFound(err error) bool {
	return errs.Unwrap(err) == mongo.ErrNoDocuments
}

func FindPageWithCursor[T any](ctx context.Context, coll *mongo.Collection, cursor int64, cursorField string, sortDirection int32, limit int64, filter bson.M, sort bson.D, pipeline mongo.Pipeline) ([]T, string, error) {

	var _pipeline []bson.D

	// 添加过滤条件
	if filter != nil {
		_pipeline = append(_pipeline, bson.D{{Key: "$match", Value: filter}})
	}

	// 添加游标条件
	if cursor != 0 {
		var comparisonOperator string
		if sortDirection >= 0 {
			comparisonOperator = "$gt"
		} else {
			comparisonOperator = "$lt"
		}
		cursorTime := time.Unix(0, cursor*int64(time.Millisecond))
		_pipeline = append(_pipeline, bson.D{{Key: "$match", Value: bson.M{cursorField: bson.M{comparisonOperator: cursorTime}}}})
	}

	_pipeline = append(_pipeline, pipeline...)

	_pipeline = append(_pipeline,
		bson.D{{Key: "$sort", Value: sort}},
		bson.D{{Key: "$limit", Value: limit}},
	)

	// 执行聚合查询
	cur, err := coll.Aggregate(ctx, _pipeline)
	if err != nil {
		return nil, "", errs.WrapMsg(err, "mongo failed to execute aggregation")
	}
	defer cur.Close(ctx)

	var results []T
	if err := cur.All(ctx, &results); err != nil {
		return nil, "", errs.WrapMsg(err, "mongo failed to decode aggregation results")
	}

	// 获取下一个游标
	var nextCursor string
	if len(results) > 0 {
		lastResult := results[len(results)-1]
		nextCursor, err = getFieldValue(lastResult, cursorField)
		if err != nil {
			return nil, "", errs.WrapMsg(err, "failed to get next cursor")
		}
	}

	return results, nextCursor, nil
}
func getFieldValue(v interface{}, field string) (string, error) {
	r := reflect.ValueOf(v)
	if r.Kind() == reflect.Ptr {
		r = r.Elem()
	}
	if r.Kind() != reflect.Struct {
		return "", fmt.Errorf("v must be a struct or pointer to struct")
	}

	f := r.FieldByName(field)
	if !f.IsValid() {
		return "", fmt.Errorf("no such field: %s in obj", field)
	}

	switch f.Kind() {
	case reflect.String:
		return f.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(f.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(f.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(f.Float(), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(f.Bool()), nil
	case reflect.Struct:
		// Handle common MongoDB types
		if f.Type().String() == "primitive.ObjectID" {
			// Use String() method if available
			if strMethod := f.MethodByName("String"); strMethod.IsValid() {
				return strMethod.Call(nil)[0].String(), nil
			}
		}
		// Handle time.Time
		if f.Type().String() == "time.Time" {
			t := f.Interface().(time.Time)
			if t.IsZero() {
				return "0", nil
			}
			return strconv.FormatInt(t.UnixMilli(), 10), nil
		}
		// Convert the struct to a string representation
		return fmt.Sprintf("%v", f.Interface()), nil
	default:
		return "", fmt.Errorf("unsupported field type: %v", f.Kind())
	}
}
