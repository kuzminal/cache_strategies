package lru

import (
	"LRU_cache/pkg/cache"
	"container/list"
)

type Item struct {
	Key   interface{}
	Value interface{}
}

type LRU struct {
	capacity int
	items    map[interface{}]*list.Element
	queue    *list.List
}

func (L *LRU) Add(key, value interface{}) bool {
	if element, exists := L.items[key]; exists == true {
		L.queue.MoveToFront(element)
		return false
	}

	if L.capacity == 0 {
		return true
	}

	if L.queue.Len() == L.capacity {
		L.removeLastElement()
	}

	item := &Item{
		Key:   key,
		Value: value,
	}

	element := L.queue.PushFront(item)
	L.items[item.Key] = element

	return true
}

func (L *LRU) Get(key interface{}) (value interface{}, ok bool) {
	element, exists := L.items[key]
	if !exists {
		return "", false
	}
	L.queue.MoveToFront(element)
	return element.Value.(*Item).Value, true
}

func (L *LRU) Remove(key interface{}) (ok bool) {
	element, exists := L.items[key]
	if exists {
		L.queue.Remove(element)
		delete(L.items, key)
		return true
	} else {
		return false
	}
}

func (L *LRU) removeLastElement() {
	if element := L.queue.Back(); element != nil {
		item := L.queue.Remove(element).(*Item)
		delete(L.items, item.Key)
	}
}

func NewLRUCache(n int) cache.Cache {
	if n <= 0 {
		panic("capacity must be positive")
	}
	return &LRU{
		capacity: n,
		items:    make(map[interface{}]*list.Element),
		queue:    list.New(),
	}
}
