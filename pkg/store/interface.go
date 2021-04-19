package store

import (
	"fmt"
	"github.com/yametech/verthandi/pkg/core"
	"github.com/yametech/verthandi/pkg/store/gtm"
)

type ErrorType error

var (
	NotFound ErrorType = fmt.Errorf("notFound")
)

type WatchInterface interface {
	ResultChan() <-chan core.IObject
	Handle(*gtm.Op) error
	ErrorStop() chan error
	CloseStop() chan struct{}
}

type Watch struct {
	r     chan core.IObject
	err   chan error
	c     chan struct{}
	coder Coder
}

type Coder interface {
	Decode(*gtm.Op) (core.IObject, error)
}

var coderList = make(map[string]Coder)

func AddResourceCoder(res string, coder Coder) {
	coderList[res] = coder
}

func GetResourceCoder(res string) Coder {
	coder, exist := coderList[res]
	if !exist {
		return nil
	}
	return coder
}

func NewWatch(coder Coder) *Watch {
	return &Watch{
		r:     make(chan core.IObject, 1),
		err:   make(chan error),
		c:     make(chan struct{}),
		coder: coder,
	}
}

// Delegate Handle
func (w *Watch) Handle(op *gtm.Op) error {
	obj, err := w.coder.Decode(op)
	if err != nil {
		return err
	}
	w.r <- obj
	return nil
}

// ResultChan
func (w *Watch) ResultChan() <-chan core.IObject {
	return w.r
}

// ErrorStop
func (w *Watch) CloseStop() chan struct{} {
	return w.c
}

// ErrorStop
func (w *Watch) ErrorStop() chan error {
	return w.err
}

type IStore interface {
	List(namespace, resource, labels string, sort map[string]interface{}, skip, limit int64) ([]interface{}, int64, error)
	ListByFilter(namespace, resource string, filter, sort map[string]interface{}, skip, limit int64) ([]interface{}, error)
	GetByUUID(namespace, resource, uuid string, result interface{}) error
	GetByFilter(namespace, resource string, result interface{}, filter map[string]interface{}) error
	Create(namespace, resource string, object core.IObject) (core.IObject, error)
	Apply(namespace, resource, uuid string, object core.IObject) (core.IObject, bool, error)
	Delete(namespace, resource, uuid string) error
	Watch2(namespace, resource string, resourceVersion int64, watch WatchInterface)
}
