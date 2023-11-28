package main

import "sync"

// Cache 定义一个基本的内存缓存结构。
type Cache struct {
	mu    sync.Mutex
	store map[string]string
}

// NewCache 创建并返回一个新的缓存实例。
func NewCache() *Cache {
	return &Cache{
		store: make(map[string]string),
	}
}

// Set 设置键值对到缓存中。
func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
}

// Get 从缓存中获取键对应的值。
func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, exists := c.store[key]
	return val, exists
}

// Delete 从缓存中删除键对应的值。
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}
