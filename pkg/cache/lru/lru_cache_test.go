package lru

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Тест: корректное добавление элемента в пустой кеш
func TestLRU_Add_NewElement(t *testing.T) {
	lru := NewLRUCache(2).(*LRU)

	ok := lru.Add("key1", "value1")

	assert.True(t, ok, "Expected Add to return true for new key")
	assert.Equal(t, 1, lru.queue.Len(), "Queue length should be 1")
	assert.Contains(t, lru.items, "key1", "Items map should contain key1")
}

// Тест: добавление существующего ключа (должен вернуть false, переместить в начало)
func TestLRU_Add_ExistingKey(t *testing.T) {
	lru := NewLRUCache(2).(*LRU)
	lru.Add("key1", "value1")

	ok := lru.Add("key1", "value2")

	assert.False(t, ok, "Expected Add to return false for existing key")
	assert.Equal(t, "value1", lru.queue.Front().Value.(*Item).Value,
		"Value should not be updated on Add for existing key")
	assert.Equal(t, 1, lru.queue.Len(), "Queue length should still be 1")
}

// Тест: превышение ёмкости (LRU-вытеснение)
func TestLRU_Add_Overflow(t *testing.T) {
	lru := NewLRUCache(2).(*LRU)
	lru.Add("key1", "value1")
	lru.Add("key2", "value2")

	// Добавляем третий — должен вытесниться key2 (т.к. он в конце)
	lru.Add("key3", "value3")

	assert.NotContains(t, lru.items, "key1", "key1 should be evicted")
	_, ok1 := lru.Get("key1")
	assert.False(t, ok1, "key1 should not be retrievable")

	_, ok2 := lru.Get("key2")
	assert.True(t, ok2, "key2 should still exist")

	_, ok3 := lru.Get("key3")
	assert.True(t, ok3, "key3 should be added")
	assert.Equal(t, 2, lru.queue.Len(), "Queue should hold exactly 2 elements")
}

// Тест: Get существующего элемента — повышение приоритета
func TestLRU_Get_Existing(t *testing.T) {
	lru := NewLRUCache(2).(*LRU)
	lru.Add("key1", "value1")
	lru.Add("key2", "value2")

	val, ok := lru.Get("key1")
	assert.True(t, ok, "key1 should be found")
	assert.Equal(t, "value1", val, "Returned value should match")

	// Проверяем, что key1 теперь в начале очереди
	frontKey := lru.queue.Front().Value.(*Item).Key
	assert.Equal(t, "key1", frontKey, "key1 should be moved to front")
}

// Тест: Get несуществующего элемента
func TestLRU_Get_NonExistent(t *testing.T) {
	lru := NewLRUCache(2).(*LRU)

	val, ok := lru.Get("unknown")
	assert.False(t, ok, "Get should return false for unknown key")
	assert.Empty(t, val, "Value should be empty")
}

// Тест: Remove существующего элемента
func TestLRU_Remove_Existing(t *testing.T) {
	lru := NewLRUCache(2).(*LRU)
	lru.Add("key1", "value1")

	ok := lru.Remove("key1")
	assert.True(t, ok, "Remove should return true for existing key")
	assert.Equal(t, 0, lru.queue.Len(), "Queue should be empty")
	assert.NotContains(t, lru.items, "key1", "Items map should not contain key1")
}

// Тест: Remove несуществующего элемента
func TestLRU_Remove_NonExistent(t *testing.T) {
	lru := NewLRUCache(2).(*LRU)

	ok := lru.Remove("unknown")
	assert.False(t, ok, "Remove should return false for non-existent key")
}

// Тест: ёмкость 0 — добавление элементов
func TestLRU_ZeroCapacity(t *testing.T) {
	lru := NewLRUCache(0).(*LRU)

	ok := lru.Add("key1", "value1")
	assert.True(t, ok, "Add should return true even with capacity 0")

	_, ok = lru.Get("key1")
	assert.False(t, ok, "No element should be stored when capacity is 0")
	assert.Equal(t, 0, lru.queue.Len(), "Queue should remain empty")
}

// Тест: ёмкость 1 — замена элемента
func TestLRU_CapacityOne_Replace(t *testing.T) {
	lru := NewLRUCache(1).(*LRU)
	lru.Add("key1", "value1")

	lru.Add("key2", "value2")

	_, ok1 := lru.Get("key1")
	assert.False(t, ok1, "key1 should be evicted")

	_, ok2 := lru.Get("key2")
	assert.True(t, ok2, "key2 should be present")
}

// Тест: Get + Add — приоритет обновляется
func TestLRU_GetThenAdd_RespectsPriority(t *testing.T) {
	lru := NewLRUCache(2).(*LRU)
	lru.Add("key1", "value1") // очередь: [key1]
	lru.Add("key2", "value2") // очередь: [key2, key1]

	lru.Get("key1") // очередь: [key1, key2]

	lru.Add("key3", "value3") // key2 должен вытесниться

	_, ok := lru.Get("key2")
	assert.False(t, ok, "key2 should be evicted")
	_, ok = lru.Get("key1")
	assert.True(t, ok, "key1 should remain due to recent access")
}

// Тест: пустые ключи и значения
func TestLRU_EmptyKeyAndValue(t *testing.T) {
	lru := NewLRUCache(2).(*LRU)

	// Пустой ключ
	lru.Add("", "value")
	val, ok := lru.Get("")
	assert.True(t, ok, "Empty key should be allowed")
	assert.Equal(t, "value", val, "Should retrieve value for empty key")

	// Пустое значение
	lru.Add("key", "")
	val, ok = lru.Get("key")
	assert.True(t, ok, "Empty value should be allowed")
	assert.Equal(t, "", val, "Should retrieve empty value")
}

// Тест: множественные операции — граничное поведение
func TestLRU_MultipleOperations(t *testing.T) {
	lru := NewLRUCache(3).(*LRU)

	lru.Add("a", "1")
	lru.Add("b", "2")
	lru.Add("c", "3") // очередь: c -> b -> a

	lru.Get("a") // a становится самым приоритетным: a -> c -> b

	lru.Add("d", "4") // вытесняется b

	_, okB := lru.Get("b")
	assert.False(t, okB, "b should be evicted")

	_, okA := lru.Get("a")
	assert.True(t, okA, "a should remain due to access")

	front := lru.queue.Front().Value.(*Item).Key
	assert.Equal(t, "a", front, "a should be at front after access")
}
