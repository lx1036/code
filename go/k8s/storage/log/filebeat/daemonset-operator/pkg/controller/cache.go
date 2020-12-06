package controller

import (
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"sync"
)

type Cache struct {
	sync.Mutex

	items map[string]*corev1.Pod

	changed bool
}

func newCache() *Cache {
	c := &Cache{
		items: make(map[string]*corev1.Pod),
	}

	return c
}

func (c *Cache) Set(key string, value *corev1.Pod) {
	c.Lock()
	defer c.Unlock()

	c.items[key] = value
	c.changed = true
}

func (c *Cache) Delete(key string) {
	c.Lock()
	defer c.Unlock()

	delete(c.items, key)
	c.changed = true
}

func (c *Cache) Get(key string) *corev1.Pod {
	c.Lock()
	defer c.Unlock()

	pod, ok := c.items[key]
	if !ok {
		log.Warnf("pod key %s is not found in cache", key)
		return nil
	}

	return pod
}
