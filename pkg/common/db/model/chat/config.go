package chat

import (
	"context"

	"github.com/openimsdk/chat/pkg/common/db/table/chat"
	"github.com/openimsdk/tools/db/mongoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewAppConfig(db *mongo.Database) (chat.AppConfigInterface, error) {
	coll := db.Collection("app_config")
	return &AppConfig{coll: coll}, nil
}

type AppConfig struct {
	coll *mongo.Collection
}

func (o *AppConfig) GetVersionConfig(ctx context.Context) (*chat.AppVersionConfig, error) {
	filter := bson.D{{Key: "name", Value: "app_version"}}
	return mongoutil.FindOne[*chat.AppVersionConfig](ctx, o.coll, filter)
}

func (o *AppConfig) GetFakeUserConfig(ctx context.Context) (*chat.AppFakeUserConfig, error) {
	filter := bson.D{{Key: "name", Value: "app_fake_user"}}
	return mongoutil.FindOne[*chat.AppFakeUserConfig](ctx, o.coll, filter)
}
