package chat

import (
	"context"
	"time"

	"github.com/openimsdk/chat/pkg/common/db/table/chat"
	"github.com/openimsdk/tools/db/mongoutil"
	"github.com/openimsdk/tools/errs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserPostRelation struct {
	coll *mongo.Collection
}

func NewUserPostRelation(db *mongo.Database) (chat.UserPostRelationInterface, error) {
	coll := db.Collection("user_post_relation")
	_, err := coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "post_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return &UserPostRelation{coll: coll}, nil
}

func (o *UserPostRelation) Create(ctx context.Context, relations []*chat.UserPostRelation) error {
	for i, relation := range relations {
		if relation.CreateTime.IsZero() {
			relations[i].CreateTime = time.Now()
		}
		if relation.UpdateTime.IsZero() {
			relations[i].UpdateTime = time.Now()
		}
	}
	return mongoutil.InsertMany(ctx, o.coll, relations)
}

func (o *UserPostRelation) Take(ctx context.Context, userID, postID string) (*chat.UserPostRelation, error) {
	return mongoutil.FindOne[*chat.UserPostRelation](ctx, o.coll, bson.M{"post_id": postID, "user_id": userID})
}

func (o *UserPostRelation) UpdateByMap(ctx context.Context, userID, postID string, args map[string]any) error {
	if len(args) == 0 {
		return nil
	}
	filter := bson.M{"post_id": postID, "user_id": userID}
	args["update_time"] = time.Now()
	return mongoutil.UpdateOne(ctx, o.coll, filter, bson.M{"$set": args}, false)
}

func (o *UserPostRelation) Delete(ctx context.Context, userID, postID string) error {
	return mongoutil.DeleteOne(ctx, o.coll, bson.M{"post_id": postID})
}

func (o *UserPostRelation) GetLikeCount(ctx context.Context, postID string) (int64, error) {
	return mongoutil.Count(ctx, o.coll, bson.M{"post_id": postID, "is_liked": 1})
}

func (o *UserPostRelation) GetCollectCount(ctx context.Context, postID string) (int64, error) {
	return mongoutil.Count(ctx, o.coll, bson.M{"post_id": postID, "is_collected": 1})
}

func (o *UserPostRelation) GetForwardCount(ctx context.Context, postID string) (int64, error) {
	return mongoutil.Count(ctx, o.coll, bson.M{"post_id": postID, "is_forwarded": 1})
}

func (o *UserPostRelation) GetCommentCount(ctx context.Context, postID string) (int64, error) {
	return mongoutil.Count(ctx, o.coll, bson.M{"post_id": postID, "is_commented": 1})
}

func (o *UserPostRelation) GetLikeCounts(ctx context.Context, postIDs []string) (map[string]int64, error) {
	cursor, err := o.coll.Find(ctx, bson.M{"post_id": bson.M{"$in": postIDs}, "is_liked": 1})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	counts := make(map[string]int64)
	for cursor.Next(ctx) {
		var result struct {
			PostID string `bson:"post_id"`
			Count  int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		counts[result.PostID] = result.Count
	}
	return counts, nil
}

func (o *UserPostRelation) GetCommentCounts(ctx context.Context, postIDs []string) (map[string]int64, error) {
	cursor, err := o.coll.Find(ctx, bson.M{"post_id": bson.M{"$in": postIDs}, "is_commented": 1})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	counts := make(map[string]int64)
	for cursor.Next(ctx) {
		var result struct {
			PostID string `bson:"post_id"`
			Count  int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		counts[result.PostID] = result.Count
	}
	return counts, nil
}

func (o *UserPostRelation) GetForwardCounts(ctx context.Context, postIDs []string) (map[string]int64, error) {
	cursor, err := o.coll.Find(ctx, bson.M{"post_id": bson.M{"$in": postIDs}, "is_forwarded": 1})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	counts := make(map[string]int64)
	for cursor.Next(ctx) {
		var result struct {
			PostID string `bson:"post_id"`
			Count  int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		counts[result.PostID] = result.Count
	}
	return counts, nil
}

func (o *UserPostRelation) GetIsLiked(ctx context.Context, userID, postID string) (int32, error) {
	return mongoutil.FindOne[int32](ctx, o.coll, bson.M{"user_id": userID, "post_id": postID, "is_liked": 1})
}

func (o *UserPostRelation) GetIsCollected(ctx context.Context, userID, postID string) (int32, error) {
	return mongoutil.FindOne[int32](ctx, o.coll, bson.M{"user_id": userID, "post_id": postID, "is_collected": 1})
}

func (o *UserPostRelation) GetIsForwarded(ctx context.Context, userID, postID string) (int32, error) {
	return mongoutil.FindOne[int32](ctx, o.coll, bson.M{"user_id": userID, "post_id": postID, "is_forwarded": 1})
}

func (o *UserPostRelation) GetIsCommented(ctx context.Context, userID, postID string) (int32, error) {
	return mongoutil.FindOne[int32](ctx, o.coll, bson.M{"user_id": userID, "post_id": postID, "is_commented": 1})
}

func (o *UserPostRelation) GetReplyPostIDs(ctx context.Context, userID string) ([]string, error) {
	return mongoutil.Find[string](ctx, o.coll, bson.M{"user_id": userID, "is_commented": 1}, options.Find().SetProjection(bson.M{"_id": 0, "post_id": 1}))
}

func (o *UserPostRelation) GetLikedPostIDs(ctx context.Context, userID string) ([]string, error) {
	return mongoutil.Find[string](ctx, o.coll, bson.M{"user_id": userID, "is_liked": 1}, options.Find().SetProjection(bson.M{"_id": 0, "post_id": 1}))
}

func (o *UserPostRelation) GetCollectedPostIDs(ctx context.Context, userID string) ([]string, error) {
	return mongoutil.Find[string](ctx, o.coll, bson.M{"user_id": userID, "is_collected": 1}, options.Find().SetProjection(bson.M{"_id": 0, "post_id": 1}))
}
