package store

import (
	"fmt"
	"github.com/yametech/echoer-flow-tool/pkg/core"
)

type ErrorType error

var (
	NotFound ErrorType = fmt.Errorf("notFound")
)

type IStore interface {
	List(namespace, resource, labels string, sort map[string]interface{}, skip, limit int64) ([]interface{}, int64, error)
	ListByFilter(namespace, resource string, filter, sort map[string]interface{}, skip, limit int64) ([]interface{}, int64, error)
	GetByUUID(namespace, resource, uuid string, result interface{}) error
	GetByFilter(namespace, resource string, result interface{}, filter map[string]interface{}) error
	Create(namespace, resource string, object core.IObject) (core.IObject, error)
	Apply(namespace, resource, name string, object core.IObject) (core.IObject, bool, error)
	Delete(namespace, resource, uuid string) error
}
