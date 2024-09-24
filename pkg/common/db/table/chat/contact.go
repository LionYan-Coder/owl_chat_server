package chat

import (
	"context"
	"time"
)

type Contact struct {
	UserID     string    `bson:"user_id"`
	GroupIDs   []string  `bson:"groups"`
	CreateTime time.Time `bson:"create_time"`
	ChangeTime time.Time `bson:"change_time"`
}

func (Contact) TableName() string {
	return "contacts"
}

type ContactInterface interface {
	AddGroup(ctx context.Context, userId string, groupIDs []string) error
	DeleteGroup(ctx context.Context, userID string, groupIDs []string) error
	TakeGroups(ctx context.Context, userId string) (*Contact, error)
}
