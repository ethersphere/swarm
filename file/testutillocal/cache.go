package testutillocal

import (
	"context"

	"github.com/ethersphere/swarm/file"
)

var (
	defaultSectionSize = 32
	defaultBranches    = 128
)

type Cache struct {
	data  map[int][]byte
	index int
	w     file.SectionWriter
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[int][]byte),
	}
}

func (c *Cache) Init(_ context.Context, _ func(error)) {
}

func (c *Cache) SetWriter(writeFunc file.SectionWriterFunc) file.SectionWriter {
	c.w = writeFunc(nil)
	return c
}

func (c *Cache) SetSpan(length int) {
	if c.w != nil {
		c.w.SetSpan(length)
	}
}

func (c *Cache) SetLength(length int) {
	if c.w != nil {
		c.w.SetLength(length)
	}
}

func (c *Cache) SeekSection(offset int) {
	c.index = offset
	if c.w != nil {
		c.w.SeekSection(offset)
	}
}

func (c *Cache) Write(b []byte) (int, error) {
	c.data[c.index] = b
	if c.w != nil {
		return c.w.Write(b)
	}
	return len(b), nil
}

func (c *Cache) Sum(b []byte) []byte {
	if c.w == nil {
		return nil
	}
	return c.w.Sum(b)
}

func (c *Cache) Reset() {
	if c.w == nil {
		return
	}
	c.w.Reset()
}

func (c *Cache) SectionSize() int {
	if c.w != nil {
		return c.w.SectionSize()
	}
	return defaultSectionSize
}

func (c *Cache) BlockSize() int {
	return c.SectionSize()
}

func (c *Cache) Size() int {
	if c.w != nil {
		return c.w.Size()
	}
	return defaultSectionSize
}

func (c *Cache) Branches() int {
	if c.w != nil {
		return c.w.Branches()
	}
	return defaultBranches
}

func (c *Cache) Get(index int) []byte {
	return c.data[index]
}

func (c *Cache) Delete(index int) {
	delete(c.data, index)
}
