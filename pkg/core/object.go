package core

import (
	"encoding/json"
	"github.com/yametech/verthandi/pkg/utils"
	"time"
)

type Kind string

type IObject interface {
	GetUUID() string
	GetKind() string
	GetName() string
	Delete()
	GenerateVersion() IObject
	GetResourceVersion() int64
	Clone() IObject
}

type Metadata struct {
	Name        string                 `json:"name" bson:"name"`
	Kind        string                 `json:"kind"  bson:"kind"`
	UUID        string                 `json:"uuid" bson:"uuid"`
	Version     int64                  `json:"version" bson:"version"`
	IsDelete    bool                   `json:"is_delete" bson:"is_delete"`
	CreatedTime int64                  `json:"created_time" bson:"created_time"`
	Labels      map[string]interface{} `json:"labels" bson:"labels"`
}

func (m *Metadata) Clone() IObject {
	panic("implement me")
}

func (m *Metadata) GetKind() string {
	return m.Kind
}

func (m *Metadata) GetName() string {
	return m.Name
}

func (m *Metadata) GenerateVersion() IObject {
	m.Version = time.Now().Unix()
	if m.UUID == "" {
		m.UUID = utils.NewSUID().String()
	}
	if m.CreatedTime == 0 {
		m.CreatedTime = time.Now().Unix()
	}
	return m
}

func (m *Metadata) GetResourceVersion() int64 {
	return m.Version
}

func (m *Metadata) GetUUID() string {
	return m.UUID
}

func (m *Metadata) Delete() {
	m.IsDelete = true
}

func Clone(src, tag interface{}) {
	b, _ := json.Marshal(src)
	_ = json.Unmarshal(b, tag)
}

func EncodeFromMap(i interface{}, m map[string]interface{}) error {
	bs, err := json.Marshal(&m)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, i); err != nil {
		return err
	}
	return nil
}

func ToMap(i interface{}) (map[string]interface{}, error) {
	var result = make(map[string]interface{})
	bs, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bs, &result); err != nil {
		return nil, err
	}
	return result, err
}

func UnmarshalInterfaceToResource(src interface{}, dest IObject) error {
	bs, err := json.Marshal(src)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, dest); err != nil {
		return err
	}
	return nil
}
