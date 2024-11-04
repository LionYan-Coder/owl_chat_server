package chat

import (
	"context"

	chatpb "github.com/openimsdk/chat/pkg/protocol/chat"
)

func (o *chatSvr) CheckVersion(ctx context.Context, req *chatpb.CheckVersionReq) (*chatpb.CheckVersionResp, error) {
	if err := req.Check(); err != nil {
		return nil, err
	}

	config, err := o.Database.GetVersionConfig(ctx)
	if err != nil {
		return nil, err
	}

	languageConfig := config.Config.Languages[req.Language]

	return &chatpb.CheckVersionResp{
		BuildVersion:                languageConfig.BuildVersion,
		DownloadURL:                 languageConfig.DownloadURL,
		BuildUpdateTitle:            languageConfig.BuildUpdateTitle,
		BuildUpdateDescriptionTitle: languageConfig.BuildUpdateDescriptionTitle,
		BuildUpdateDescriptions:     languageConfig.BuildUpdateDescriptions,
		NeedForceUpdate:             languageConfig.NeedForceUpdate,
	}, nil
}

func (o *chatSvr) GetFakeUser(ctx context.Context, req *chatpb.GetFakeUserReq) (*chatpb.GetFakeUserResp, error) {
	config, err := o.Database.GetFakeUserConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &chatpb.GetFakeUserResp{
		Online: int32(config.Config.Online),
		Total:  int32(config.Config.Total),
	}, nil
}
