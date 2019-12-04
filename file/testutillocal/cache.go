package testutillocal

import (
	"context"

	"github.com/ethersphere/swarm/param"
)

type Cache struct {
	data map[int][]byte
	w    param.SectionWriter
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[int][]byte),
	}
}

func (c *Cache) Init(_ context.Context, _ func(error)) {
}

func (c *Cache) Link(writeFunc func() param.SectionWriter) {
	c.w = writeFunc()
}

func (c *Cache) Write(index int, b []byte) {
	c.data[index] = b
	if c.w == nil {
		return
	}
	c.w.Write(index, b)
}

func (c *Cache) Sum(b []byte, length int, span []byte) []byte {
	if c.w == nil {
		return nil
	}
	return c.w.Sum(b, length, span)
}

func (c *Cache) Reset(ctx context.Context) {
	if c.w == nil {
		return
	}
	c.w.Reset(ctx)
}

func (c *Cache) Get(index int) []byte {
	return c.data[index]
}
