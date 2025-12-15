package lfu

import (
	"container/list"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewLFUCache_InvalidCapacity проверяет создание кэша с некорректной ёмкостью
func TestNewLFUCache_InvalidCapacity(t *testing.T) {
	assert.Panics(t, func() {
		NewLFUCache(0)
	}, "Expected panic for capacity <= 0")

	assert.Panics(t, func() {
		NewLFUCache(-1)
	}, "Expected panic for negative capacity")
}

// TestNewLFUCache_ValidCapacity проверяет создание кэша с корректной ёмкостью
func TestNewLFUCache_ValidCapacity(t *testing.T) {
	cache := NewLFUCache(3)
	assert.Equal(t, 3, cache.capacity, "Capacity should be set correctly")
	assert.Equal(t, 0, cache.minFreq, "Initial minFreq should be 0")
	assert.NotNil(t, cache.items, "Items map should be initialized")
	assert.NotNil(t, cache.freqLists, "FreqLists map should be initialized")
	assert.Equal(t, 0, cache.freqNodes.Len(), "FreqNodes list should be empty")
}

// TestGet_NonExistentKey проверяет получение несуществующего ключа
func TestGet_NonExistentKey(t *testing.T) {
	cache := NewLFUCache(2)
	val, ok := cache.Get("unknown")
	assert.Nil(t, val, "Value should be nil")
	assert.False(t, ok, "Get should return false")
}

// TestPut_Get_Simple проверяет простое добавление и получение
func TestPut_Get_Simple(t *testing.T) {
	cache := NewLFUCache(2)
	cache.Put("key1", "value1")

	val, ok := cache.Get("key1")
	assert.True(t, ok, "Key should exist")
	assert.Equal(t, "value1", val, "Returned value should match")
	assert.Equal(t, 1, cache.minFreq, "minFreq should be 1 after first insert")
}

// TestPut_ExistingKey обновляет существующий ключ
func TestPut_ExistingKey(t *testing.T) {
	cache := NewLFUCache(2)
	cache.Put("key1", "value1")
	cache.Put("key1", "value2")

	val, ok := cache.Get("key1")
	assert.True(t, ok, "Key should exist")
	assert.Equal(t, "value2", val, "Value should be updated")
	assert.Equal(t, 1, cache.minFreq, "minFreq should still be 1 (same frequency)")
}

// TestPut_TriggerIncrementFrequency проверяет, что Get увеличивает частоту
func TestPut_TriggerIncrementFrequency(t *testing.T) {
	cache := NewLFUCache(2)
	cache.Put("key1", "value1")
	cache.Get("key1") // Частота станет 2
	cache.Get("key1") // Частота станет 3

	elem := cache.items["key1"]
	item := elem.Value.(*CacheItem)
	assert.Equal(t, 3, item.frequency, "Frequency should be incremented to 3")
}

// TestEvict_LFU_Basic проверяет базовое вытеснение LFU элемента
func TestEvict_LFU_Basic(t *testing.T) {
	cache := NewLFUCache(2)
	cache.Put("key1", "value1") // freq=1
	cache.Put("key2", "value2") // freq=1
	cache.Get("key1")           // key1.freq=2, key2.freq=1 -> key2 будет вытеснён

	cache.Put("key3", "value3") // Должен вытеснить key2

	_, ok1 := cache.Get("key1")
	assert.True(t, ok1, "key1 should remain")

	_, ok2 := cache.Get("key2")
	assert.False(t, ok2, "key2 should be evicted")

	_, ok3 := cache.Get("key3")
	assert.True(t, ok3, "key3 should be added")
}

// TestEvict_LFU_SameFreq_LRU проверяет, что при одинаковой частоте вытесняется LRU
func TestEvict_LFU_SameFreq_LRU(t *testing.T) {
	cache := NewLFUCache(2)
	cache.Put("key1", "value1") // freq=1, order: key1
	cache.Put("key2", "value2") // freq=1, order: key2 -> key1

	cache.Put("key3", "value3") // Оба с freq=1, вытеснится key1 (LRU)

	_, ok1 := cache.Get("key1")
	assert.False(t, ok1, "key1 (LRU) should be evicted")

	_, ok2 := cache.Get("key2")
	assert.True(t, ok2, "key2 should remain")

	_, ok3 := cache.Get("key3")
	assert.True(t, ok3, "key3 should be added")
}

// TestEvict_UpdateMinFreq проверяет обновление minFreq при удалении последнего элемента
func TestEvict_UpdateMinFreq(t *testing.T) {
	cache := NewLFUCache(2)
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	cache.Get("key1") // freq: key1=2, key2=1

	cache.Put("key3", "value3") // evict key2 -> minFreq должен стать 2
	assert.Equal(t, 2, cache.minFreq, "minFreq should update to 2 after evicting last freq=1")
}

// TestSize проверяет корректность подсчёта размера
func TestSize(t *testing.T) {
	cache := NewLFUCache(3)
	assert.Equal(t, 0, cache.Size(), "Initial size should be 0")

	cache.Put("k1", "v1")
	cache.Put("k2", "v2")
	assert.Equal(t, 2, cache.Size(), "Size should be 2")

	cache.Get("k1")
	assert.Equal(t, 2, cache.Size(), "Size should remain 2 after Get")

	cache.Put("k3", "v3")
	cache.Put("k4", "v4") // evict one
	assert.Equal(t, 3, cache.Size(), "Size should be capped at capacity")
}

// TestClear очищает кэш
func TestClear(t *testing.T) {
	cache := NewLFUCache(2)
	cache.Put("key1", "value1")
	cache.Clear()

	assert.Equal(t, 0, cache.Size(), "Size should be 0 after Clear")
	assert.Equal(t, 0, cache.minFreq, "minFreq should be 0")
	assert.Equal(t, 0, cache.freqNodes.Len(), "freqNodes should be empty")
	_, ok := cache.Get("key1")
	assert.False(t, ok, "No keys should exist after Clear")
}

// TestPut_ZeroCapacity игнорирует операции при ёмкости 0 (но мы не создаём такие)
func TestPut_ZeroCapacity(t *testing.T) {
	// Наша реализация паникует при capacity <= 0, так что тестируем только поведение при 1
	cache := NewLFUCache(1)
	cache.Put("key1", "value1")
	cache.Put("key2", "value2") // evict key1

	_, ok := cache.Get("key1")
	assert.False(t, ok, "key1 should be evicted")
	assert.Equal(t, 1, cache.Size(), "Size should remain 1")
}

// TestFrequencyNode_Insertion проверяет корректность вставки узлов частот
func TestFrequencyNode_Insertion(t *testing.T) {
	cache := NewLFUCache(3)

	// Вставляем freq=3
	node3 := &FrequencyNode{freq: 3, elements: list.New()}
	cache.insertFrequencyNode(3, node3)
	assert.Equal(t, 1, cache.freqNodes.Len(), "Should have one node")

	// Вставляем freq=1 — должен быть в начале
	node1 := &FrequencyNode{freq: 1, elements: list.New()}
	cache.insertFrequencyNode(1, node1)
	assert.Equal(t, 2, cache.freqNodes.Len(), "Should have two nodes")
	assert.Equal(t, 1, cache.freqNodes.Front().Value.(*FrequencyNode).freq, "Freq=1 should be first")

	// Вставляем freq=2 — должен встать посередине
	node2 := &FrequencyNode{freq: 2, elements: list.New()}
	cache.insertFrequencyNode(2, node2)
	assert.Equal(t, 3, cache.freqNodes.Len(), "Three nodes expected")
	e := cache.freqNodes.Front()
	assert.Equal(t, 1, e.Value.(*FrequencyNode).freq)
	e = e.Next()
	assert.Equal(t, 2, e.Value.(*FrequencyNode).freq)
	e = e.Next()
	assert.Equal(t, 3, e.Value.(*FrequencyNode).freq)
}

// TestRemoveFromFrequencyList проверяет удаление элемента из списка частот
func TestRemoveFromFrequencyList(t *testing.T) {
	cache := NewLFUCache(2)
	item := &CacheItem{key: "k", value: "v", frequency: 5}
	elem := cache.addToFrequencyList(5, item)

	assert.NotNil(t, cache.getFrequencyList(5), "List for freq=5 should exist")
	assert.Equal(t, 1, cache.getFrequencyList(5).Len(), "List should have one element")

	cache.removeFromFrequencyList(5, elem)
	assert.Nil(t, cache.getFrequencyList(5), "List for freq=5 should be removed")
	_, exists := cache.freqLists[5]
	assert.False(t, exists, "freqLists should not contain freq=5")
}

// TestGet_AfterEvict проверяет, что вытеснённый элемент не возвращается
func TestGet_AfterEvict(t *testing.T) {
	cache := NewLFUCache(1)
	cache.Put("k", "v")
	cache.Put("k2", "v2") // evicts k

	val, ok := cache.Get("k")
	assert.Nil(t, val)
	assert.False(t, ok, "Evicted key should not be retrievable")
}
