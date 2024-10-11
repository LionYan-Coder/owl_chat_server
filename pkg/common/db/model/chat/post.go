package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/openimsdk/chat/pkg/common/db/dbutil"
	"github.com/openimsdk/chat/pkg/common/db/table/chat"
	"github.com/openimsdk/chat/pkg/common/mctx"
	"github.com/openimsdk/tools/db/mongoutil"
	"github.com/openimsdk/tools/db/pagination"
	"github.com/openimsdk/tools/errs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewPost(db *mongo.Database) (chat.PostInterface, error) {
	coll := db.Collection("post")
	_, err := coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "post_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return &Post{coll: coll}, nil
}

type Post struct {
	coll *mongo.Collection
}

func (o *Post) Create(ctx context.Context, posts []*chat.PostDB) error {
	for i, post := range posts {
		if post.CreateTime.IsZero() {
			posts[i].CreateTime = time.Now()
		}
		if post.UpdateTime.IsZero() {
			posts[i].UpdateTime = time.Now()
		}
	}
	return mongoutil.InsertMany(ctx, o.coll, posts)
}

func (o *Post) Delete(ctx context.Context, postID string) error {
	return mongoutil.DeleteOne(ctx, o.coll, bson.M{"post_id": postID})
}
func (o *Post) Take(ctx context.Context, postID string) (*chat.Post, error) {
	pipeline := GetAggregationPipeline(ctx)
	pipeline = append([]bson.D{{{"$match", bson.M{"post_id": postID}}}}, pipeline...)

	results, err := mongoutil.Aggregate[*chat.Post](ctx, o.coll, pipeline)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	return results[0], nil
}

func (o *Post) UpdateByMap(ctx context.Context, postID string, data map[string]any) error {
	if len(data) == 0 {
		return nil
	}
	filter := bson.M{"post_id": postID}
	data["update_time"] = time.Now()
	return mongoutil.UpdateOne(ctx, o.coll, filter, bson.M{"$set": data}, false)
}

func (o *Post) PageGet(ctx context.Context, pagination pagination.Pagination) (int64, []*chat.Post, error) {
	return dbutil.FindPageWithAggregation[*chat.Post](ctx, o.coll, bson.M{}, pagination, GetAggregationPipeline(ctx))
}

func (o *Post) PageGetByUser(ctx context.Context, userID string, pagination pagination.Pagination) (int64, []*chat.Post, error) {

	return dbutil.FindPageWithAggregation[*chat.Post](ctx, o.coll, bson.M{"user_id": userID}, pagination, GetAggregationPipeline(ctx))
}

func (o *Post) PageGetByPostIDs(ctx context.Context, postIDs []string, pagination pagination.Pagination) (int64, []*chat.Post, error) {
	return dbutil.FindPageWithAggregation[*chat.Post](ctx, o.coll, bson.M{"post_id": bson.M{"$in": postIDs}}, pagination, GetAggregationPipeline(ctx))
}

func GetAggregationPipeline(ctx context.Context) mongo.Pipeline {
	opUserID, _ := mctx.CheckUser(ctx)
	return mongo.Pipeline{
		lookupUserInfo(),
		unwindUserInfo(),
		lookupRelations(),
		lookupAtUserInfo(),
		addFields(opUserID),
	}
}

func lookupUserInfo() bson.D {
	return bson.D{
		{"$lookup", bson.D{
			{"from", "attribute"},
			{"localField", "user_id"},
			{"foreignField", "user_id"},
			{"as", "user_info"},
		}},
	}
}

func unwindUserInfo() bson.D {
	return bson.D{
		{"$unwind", bson.D{
			{"path", "$user_info"},
			{"preserveNullAndEmptyArrays", true},
		}},
	}
}

func lookupRelations() bson.D {
	return bson.D{
		{"$lookup", bson.D{
			{"from", "user_post_relation"},
			{"localField", "post_id"},
			{"foreignField", "post_id"},
			{"as", "relations"},
		}},
	}
}

func lookupAtUserInfo() bson.D {
	return bson.D{
		{"$lookup", bson.D{
			{"from", "attribute"},
			{"let", bson.D{{"atUserIds", "$at_user_ids"}}},
			{"pipeline", bson.A{
				bson.D{
					{"$match", bson.D{
						{"$expr", bson.D{
							{"$cond", bson.A{
								bson.D{{"$ne", bson.A{"$$atUserIds", nil}}},
								bson.D{{"$in", bson.A{"$user_id", "$$atUserIds"}}},
								false,
							}},
						}},
					}},
				},
				bson.D{
					{"$project", bson.D{
						{"user_id", 1},
						{"nickname", 1},
						{"account", 1},
						{"address", 1},
						{"face_url", 1},
					}},
				},
			}},
			{"as", "at_user_info_list"},
		}},
	}
}

func addFields(opUserID string) bson.D {
	return bson.D{
		{"$addFields", bson.D{
			{"like_count", getCountField("is_liked")},
			{"collect_count", getCountField("is_collected")},
			{"forward_count", getCountField("is_forwarded")},
			{"is_liked", getIsField("is_liked", opUserID)},
			{"is_collected", getIsField("is_collected", opUserID)},
			{"is_forwarded", getIsField("is_forwarded", opUserID)},
		}},
	}
}

func getCountField(fieldName string) bson.D {
	return bson.D{
		{"$size", bson.D{
			{"$filter", bson.D{
				{"input", "$relations"},
				{"as", "relation"},
				{"cond", bson.D{
					{"$eq", bson.A{fmt.Sprintf("$$relation.%s", fieldName), 1}},
				}},
			}},
		}},
	}
}

func getIsField(fieldName string, opUserID string) bson.D {
	return bson.D{
		{"$cond", bson.D{
			{"if", bson.D{
				{"$gt", bson.A{
					bson.D{{"$size", bson.D{
						{"$filter", bson.D{
							{"input", "$relations"},
							{"as", "relation"},
							{"cond", bson.D{
								{"$and", bson.A{
									bson.D{{"$eq", bson.A{"$$relation.user_id", opUserID}}},
									bson.D{{"$eq", bson.A{"$$relation.post_id", "$post_id"}}},
									bson.D{{"$eq", bson.A{fmt.Sprintf("$$relation.%s", fieldName), 1}}},
								}},
							}},
						}},
					}}},
					0,
				}},
			}},
			{"then", 1},
			{"else", 0},
		}},
	}
}
