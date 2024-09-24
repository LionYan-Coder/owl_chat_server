package chat

import (
	"context"
	"errors"
	"time"

	"github.com/openimsdk/tools/db/mongoutil"
	"github.com/openimsdk/tools/errs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/openimsdk/chat/pkg/common/db/table/chat"
)

func NewContact(db *mongo.Database) (chat.ContactInterface, error) {
	coll := db.Collection("contact")
	_, err := coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return &Contact{coll: coll}, nil
}

type Contact struct {
	coll *mongo.Collection
}

// Add implements chat.ContactInterface.
func (c *Contact) AddGroup(ctx context.Context, userID string, groupIDs []string) error {
	filter := bson.M{"user_id": userID} //
	update := bson.M{
		"$setOnInsert": bson.M{"create_time": time.Now()},
		"$addToSet":    bson.M{"groups": bson.M{"$each": groupIDs}}, // 增量添加群组
		"$set":         bson.M{"change_time": time.Now()},
	}
	opts := options.Update().SetUpsert(true)
	_, err := c.coll.UpdateOne(ctx, filter, update, opts)
	return err
}

// DeleteGroup implements chat.ContactInterface.
func (c *Contact) DeleteGroup(ctx context.Context, userID string, groupIDs []string) error {
	if len(groupIDs) == 0 {
		return nil
	}

	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$pull": bson.M{"groups": bson.M{"$in": groupIDs}},
		"$set":  bson.M{"change_time": time.Now()},
	}
	result, err := c.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return errors.New("no groups were deleted or user not found")
	}
	return nil
}

// Take implements chat.ContactInterface.
func (c *Contact) TakeGroups(ctx context.Context, userId string) (*chat.Contact, error) {
	return mongoutil.FindOne[*chat.Contact](ctx, c.coll, bson.M{"user_id": userId})
}
