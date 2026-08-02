package app

import (
	"context"
	"errors"
	"sync"
)

type closeFn func(ctx context.Context) error

type closer struct {
	closes []closeFn
	mutex  sync.Mutex
}

func (c *closer) all(ctx context.Context) error {
	c.mutex.Lock()
	list := make([]closeFn, len(c.closes))
	copy(list, c.closes)
	c.closes = []closeFn{}
	c.mutex.Unlock()

	var errs []error
	for _, cl := range list {
		if err := cl(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (c *closer) add(cl closeFn) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.closes = append(c.closes, cl)
}

var globalCloser closer

func Add(cl closeFn) {
	globalCloser.add(cl)
}

func CloseAll(ctx context.Context) error {
	return globalCloser.all(ctx)
}
