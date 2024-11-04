package chat

import "context"

type LanguageConfig struct {
	NeedForceUpdate             bool     `json:"needForceUpdate"`
	DownloadURL                 string   `json:"downloadURL"`
	BuildVersion                string   `json:"buildVersion"`
	BuildUpdateTitle            string   `json:"buildUpdateTitle"`
	BuildUpdateDescriptionTitle string   `json:"buildUpdateDescriptionTitle"`
	BuildUpdateDescriptions     []string `json:"buildUpdateDescriptions"`
}

type VersionConfig struct {
	Languages map[string]LanguageConfig `json:"languages"`
}

type AppVersionConfig struct {
	Name   string        `bson:"name"`
	Config VersionConfig `bson:"config"`
}

type FakeUserConfig struct {
	Online int `json:"online"`
	Total  int `json:"total"`
}

type AppFakeUserConfig struct {
	Name   string         `bson:"name"`
	Config FakeUserConfig `bson:"config"`
}

type AppConfigInterface interface {
	GetVersionConfig(ctx context.Context) (*AppVersionConfig, error)
	GetFakeUserConfig(ctx context.Context) (*AppFakeUserConfig, error)
}
