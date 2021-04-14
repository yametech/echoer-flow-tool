package base

import (
	"github.com/yametech/echoer-flow-tool/pkg/core"
)

type ArtifactStatus uint8

const (
	Created ArtifactStatus = iota
	Building
	Built
)

type ArtifactSpec struct {
	ArtifactStatus `json:"artifact_status" bson:"artifact_status"`
	GitUrl         string `json:"git_url" bson:"git_url"`
	Language       string `json:"language" bson:"language"`
	Tag            string `json:"tag" bson:"tag" `
	Images         string `json:"images" bson:"images"`
	AppConfig      string `json:"app_config" bson:"app_config"` // gitUrl和Language存在全局配置中，不需要保存在这里
	AppName        string `json:"app_name" bson:"app_name"`     // 只存英文名，appCode不需要，用name搜索
	CreateUserId   string `json:"create_user_id" bson:"create_user_id"`
}

type Artifact struct {
	core.Metadata `json:"metadata"`
	Spec          ArtifactSpec `json:"spec"`
}

func (ar *Artifact) Clone() core.IObject {
	result := &Artifact{}
	core.Clone(ar, result)
	return result
}
