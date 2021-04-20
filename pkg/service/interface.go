package service

import (
	"fmt"
	"github.com/yametech/verthandi/pkg/core"
	"github.com/yametech/verthandi/pkg/store"
	"strconv"
)

type IService interface {
	List(namespace, resource, labels string, sort map[string]interface{}, skip, limit int64) ([]interface{}, error)
	ListByFilter(namespace, resource string, filter, sort map[string]interface{}, skip, limit int64) ([]interface{}, error)
	GetByUUID(namespace, resource, uuid string, result interface{}) error
	GetByFilter(namespace, resource string, result interface{}, filter map[string]interface{}) error
	Create(namespace, resource string, object core.IObject) (core.IObject, error)
	Apply(namespace, resource, uuid string, object core.IObject) (core.IObject, bool, error)
	Delete(namespace, resource, uuid string) error
	Watch(namespace, resource, kind, version string, objectChan chan core.IObject, closed chan struct{})
	Watch2(namespace, resource string, resourceVersion int64, watch store.WatchInterface)
}

type BaseService struct {
	store.IStore
}

func (bs *BaseService) Watch(namespace, resource, kind, version string, objectChan chan core.IObject, closed chan struct{}) {
	go func(versionStr string) {
		version, err := strconv.ParseInt(versionStr, 10, 64)
		if err != nil {
			return
		}
		coder := store.GetResourceCoder(kind)
		if coder == nil {
			return
		}
		wc := store.NewWatch(coder)
		bs.Watch2(namespace, resource, version, wc)
		for {
			select {
			case <-closed:
				return
			case err := <-wc.ErrorStop():
				fmt.Printf("user watch version: (%d) get error: (%s)\n", version, err)
				close(objectChan)
				return
			case item, ok := <-wc.ResultChan():
				if !ok {
					return
				}
				objectChan <- item
			}
		}
	}(version)
}

func NewBaseService(s store.IStore) IService {
	return &BaseService{s}
}
