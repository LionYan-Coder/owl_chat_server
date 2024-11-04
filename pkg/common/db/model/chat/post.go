package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/openimsdk/chat/pkg/common/constant"
	"github.com/openimsdk/chat/pkg/common/db/dbutil"
	"github.com/openimsdk/chat/pkg/common/db/table/chat"
	"github.com/openimsdk/chat/pkg/common/mctx"
	"github.com/openimsdk/tools/db/mongoutil"
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

func (g *Post) sortByCreateTime() bson.D {
	return bson.D{{Key: "create_time", Value: -1}}
}

func (o *Post) sortByPinedAndCreateTime() bson.D {
	return bson.D{
		{Key: "is_pined", Value: -1},
		{Key: "create_time", Value: -1},
	}
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

func (o *Post) Delete(ctx context.Context, postIDs []string) error {
	if len(postIDs) == 0 {
		return nil
	}
	return mongoutil.DeleteMany(ctx, o.coll, bson.M{"post_id": bson.M{"$in": postIDs}})
}
func (o *Post) Take(ctx context.Context, postID string) (*chat.Post, error) {
	filter := bson.M{"post_id": postID}
	results, err := mongoutil.Aggregate[*chat.Post](ctx, o.coll, GetAggregationPipeline(ctx, filter))
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

func (o *Post) GetPostsByCursorAndUserIDs(ctx context.Context, cursor int64, userIDs []string, count int64) ([]*chat.Post, string, error) {
	filter := bson.M{
		"user_id": bson.M{"$in": userIDs},
		"$or": []bson.M{
			{"comment_post_id": nil},
			{"comment_post_id": ""},
			{"comment_post_id": bson.M{"$exists": false}},
		},
	}
	sort := o.sortByCreateTime()
	return dbutil.FindPageWithCursor[*chat.Post](ctx, o.coll, cursor, "CreateTime", -1, count, filter, sort, GetAggregationPipeline(ctx))
}

func (o *Post) GetPostsByCursorAndUser(ctx context.Context, cursor int64, userID string, count int64) ([]*chat.Post, string, error) {
	filter := bson.M{"user_id": userID}
	sort := o.sortByPinedAndCreateTime()
	return dbutil.FindPageWithCursor[*chat.Post](ctx, o.coll, cursor, "CreateTime", -1, count, filter, sort, GetAggregationPipeline(ctx))
}

func (o *Post) GetPostsByCursorAndPostIDs(ctx context.Context, cursor int64, postIDs []string, count int64) ([]*chat.Post, string, error) {
	filter := bson.M{"post_id": bson.M{"$in": postIDs}}
	sort := o.sortByCreateTime()
	return dbutil.FindPageWithCursor[*chat.Post](ctx, o.coll, cursor, "CreateTime", -1, count, filter, sort, GetAggregationPipeline(ctx))
}

func (o *Post) GetCommentPostsByPostID(ctx context.Context, cursor int64, postID string, count int64) ([]*chat.Post, string, error) {
	filter := bson.M{"comment_post_id": postID}
	sort := o.sortByCreateTime()
	return dbutil.FindPageWithCursor[*chat.Post](ctx, o.coll, cursor, "CreateTime", -1, count, filter, sort, GetAggregationPipeline(ctx))
}

func (o *Post) GetCommentPostIDsByUser(ctx context.Context, userID string) ([]string, error) {
	filter := bson.M{"user_id": userID, "comment_post_id": bson.M{
		"$exists": true,
		"$nin":    []interface{}{nil, ""},
	}}
	return mongoutil.Find[string](ctx, o.coll, filter, options.Find().SetProjection(bson.M{"post_id": 1, "_id": 0}))
}

func (o *Post) GetPinnedPostByUserID(ctx context.Context, userID string) (*chat.Post, error) {
	filter := bson.M{"user_id": userID, "is_pined": constant.Pinned}
	return mongoutil.FindOne[*chat.Post](ctx, o.coll, filter)
}

// 通过forward_post_id获取post
func (o *Post) GetPostByForwardPostID(ctx context.Context, userID, forwardPostID string) (*chat.Post, error) {
	filter := bson.M{"user_id": userID, "forward_post_id": forwardPostID}
	return mongoutil.FindOne[*chat.Post](ctx, o.coll, filter)
}

func (o *Post) GetFollowedUserIDs(ctx context.Context, userID string) ([]string, error) {
	userRelationColl := o.coll.Database().Collection("friend_relation")

	filter := bson.M{"owner_user_id": userID, "is_following": 1, "is_blocked": 0}

	cursor, err := userRelationColl.Find(ctx, filter, options.Find().SetProjection(bson.M{"related_user_id": 1, "_id": 0}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		RelatedUserID string `bson:"related_user_id"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	followedUserIDs := make([]string, len(results))
	for i, result := range results {
		followedUserIDs[i] = result.RelatedUserID
	}

	return followedUserIDs, nil
}

func (o *Post) GetSubscriberUserIDs(ctx context.Context, userID string) ([]string, error) {
	// 假设user_relation集合在同一个数据库中
	userRelationColl := o.coll.Database().Collection("friend_relation")

	filter := bson.M{"owner_user_id": userID, "is_subscribed": 1, "is_blocked": 0}

	cursor, err := userRelationColl.Find(ctx, filter, options.Find().SetProjection(bson.M{"related_user_id": 1, "_id": 0}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		RelatedUserID string `bson:"related_user_id"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	subscriberUserIDs := make([]string, len(results))
	for i, result := range results {
		subscriberUserIDs[i] = result.RelatedUserID
	}

	return subscriberUserIDs, nil

}

func GetAggregationPipeline(ctx context.Context, filter ...bson.M) mongo.Pipeline {
	opUserID, _ := mctx.CheckUser(ctx)
	var _pipeline []bson.D
	if len(filter) > 0 {
		_pipeline = append(_pipeline, bson.D{{Key: "$match", Value: filter[0]}})
	}
	_pipeline = append(_pipeline,
		lookupUserInfo(),
		unwindUserInfo(),
		lookupRelations(),
		lookupForwardPost(opUserID, 3),
		unwindForwardPost(),
		lookupCommentPost(opUserID, 3),
		unwindCommentPost(),
		lookupRefPost(opUserID, 3),
		unwindRefPost(),
		lookupAtUserInfo(),
		addFields(opUserID),
	)

	return _pipeline
}

func lookupForwardPost(opUserID string, depth int) bson.D {
	if depth <= 0 {
		return bson.D{{"$addFields", bson.D{{"forward_post", nil}}}}
	}
	return bson.D{
		{"$lookup", bson.D{
			{"from", "post"},
			{"let", bson.D{{"forwardPostId", "$forward_post_id"}}},
			{"pipeline", bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$eq", bson.A{"$post_id", "$$forwardPostId"}}}}}}},
				lookupUserInfo(),
				unwindUserInfo(),
				lookupRelations(),
				lookupForwardPost(opUserID, depth-1),
				unwindForwardPost(),
				lookupCommentPost(opUserID, depth-1),
				unwindCommentPost(),
				lookupRefPost(opUserID, depth-1),
				unwindRefPost(),
				lookupAtUserInfo(),
				addFields(opUserID),
			}},
			{"as", "forward_post"},
		}},
	}
}

func lookupCommentPost(opUserID string, depth int) bson.D {
	if depth <= 0 {
		return bson.D{{"$addFields", bson.D{{"comment_post", nil}}}}
	}
	return bson.D{
		{"$lookup", bson.D{
			{"from", "post"},
			{"let", bson.D{{"commentPostId", "$comment_post_id"}}},
			{"pipeline", bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$eq", bson.A{"$post_id", "$$commentPostId"}}}}}}},
				lookupUserInfo(),
				unwindUserInfo(),
				lookupRelations(),
				lookupForwardPost(opUserID, depth-1),
				unwindForwardPost(),
				lookupCommentPost(opUserID, depth-1),
				unwindCommentPost(),
				lookupRefPost(opUserID, depth-1),
				unwindRefPost(),
				lookupAtUserInfo(),
				addFields(opUserID),
			}},
			{"as", "comment_post"},
		}},
	}
}

func lookupRefPost(opUserID string, depth int) bson.D {
	if depth <= 0 {
		return bson.D{{"$addFields", bson.D{{"ref_post", nil}}}}
	}
	return bson.D{
		{"$lookup", bson.D{
			{"from", "post"},
			{"let", bson.D{{"refPostId", "$ref_post_id"}}},
			{"pipeline", bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$eq", bson.A{"$post_id", "$$refPostId"}}}}}}},
				lookupUserInfo(),
				unwindUserInfo(),
				lookupRelations(),
				lookupForwardPost(opUserID, depth-1),
				unwindForwardPost(),
				lookupCommentPost(opUserID, depth-1),
				unwindCommentPost(),
				lookupRefPost(opUserID, depth-1),
				unwindRefPost(),
				lookupAtUserInfo(),
				addFields(opUserID),
			}},
			{"as", "ref_post"},
		}},
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

func unwindForwardPost() bson.D {
	return bson.D{
		{"$unwind", bson.D{
			{"path", "$forward_post"},
			{"preserveNullAndEmptyArrays", true},
		}},
	}
}

func unwindCommentPost() bson.D {
	return bson.D{
		{"$unwind", bson.D{
			{"path", "$comment_post"},
			{"preserveNullAndEmptyArrays", true},
		}},
	}
}

func unwindRefPost() bson.D {
	return bson.D{
		{"$unwind", bson.D{
			{"path", "$ref_post"},
			{"preserveNullAndEmptyArrays", true},
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
			{"comment_count", getCountField("is_commented")},
			{"forward_count", getCountField("is_forwarded")},
			{"is_liked", getIsField("is_liked", opUserID)},
			{"is_collected", getIsField("is_collected", opUserID)},
			{"is_commented", getIsField("is_commented", opUserID)},
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
