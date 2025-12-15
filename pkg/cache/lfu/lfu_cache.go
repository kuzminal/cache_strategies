package lfu

import (
	"container/list"
)

// CacheItem - элемент кэша
type CacheItem struct {
	key       interface{}
	value     interface{}
	frequency int // частота использования
}

// FrequencyNode - узел частоты, содержащий элементы с одной частотой
type FrequencyNode struct {
	freq     int
	elements *list.List // двусвязный список элементов с этой частотой
}

// LFUCache - основной кэш
type LFUCache struct {
	capacity int
	minFreq  int // минимальная текущая частота

	// Хранилища:
	items     map[interface{}]*list.Element // key -> элемент в elements списке
	freqLists map[int]*list.Element         // freq -> FrequencyNode в freqNodes
	freqNodes *list.List                    // список FrequencyNode, отсортированный по частоте
}

// NewLFUCache создает новый LFU кэш
func NewLFUCache(capacity int) *LFUCache {
	if capacity <= 0 {
		panic("capacity must be positive")
	}
	return &LFUCache{
		capacity:  capacity,
		minFreq:   0,
		items:     make(map[interface{}]*list.Element),
		freqLists: make(map[int]*list.Element),
		freqNodes: list.New(),
	}
}

// Get получает значение по ключу
func (c *LFUCache) Get(key interface{}) (interface{}, bool) {
	if elem, ok := c.items[key]; ok {
		// Обновляем частоту использования
		c.incrementFrequency(elem)
		item := elem.Value.(*CacheItem)
		return item.value, true
	}
	return nil, false
}

// Put добавляет или обновляет значение
func (c *LFUCache) Put(key, value interface{}) {
	if c.capacity == 0 {
		return
	}

	// Если ключ уже существует, обновляем значение и частоту
	if elem, ok := c.items[key]; ok {
		item := elem.Value.(*CacheItem)
		item.value = value
		c.incrementFrequency(elem)
		return
	}

	// Если достигли capacity, удаляем LFU элемент
	if len(c.items) >= c.capacity {
		c.evict()
	}

	// Создаем новый элемент с частотой 1
	item := &CacheItem{
		key:       key,
		value:     value,
		frequency: 1,
	}

	// Добавляем в список частоты 1
	elem := c.addToFrequencyList(1, item)
	c.items[key] = elem

	// Обновляем minFreq
	c.minFreq = 1
}

// incrementFrequency увеличивает частоту элемента
func (c *LFUCache) incrementFrequency(elem *list.Element) {
	item := elem.Value.(*CacheItem)
	oldFreq := item.frequency
	newFreq := oldFreq + 1

	// Удаляем из старого списка частот
	c.removeFromFrequencyList(oldFreq, elem)

	// Добавляем в новый список частот
	newElem := c.addToFrequencyList(newFreq, item)
	c.items[item.key] = newElem

	// Обновляем item
	updatedItem := newElem.Value.(*CacheItem)
	updatedItem.frequency = newFreq

	// Если старый список частот пуст и это была minFreq, обновляем minFreq
	if oldFreq == c.minFreq && c.getFrequencyList(oldFreq).Len() == 0 {
		c.minFreq = newFreq
	}
}

// addToFrequencyList добавляет элемент в список заданной частоты
func (c *LFUCache) addToFrequencyList(freq int, item *CacheItem) *list.Element {
	// Ищем или создаем FrequencyNode для этой частоты
	var freqNodeElem *list.Element
	if elem, ok := c.freqLists[freq]; ok {
		freqNodeElem = elem
	} else {
		// Создаем новую FrequencyNode
		freqNode := &FrequencyNode{
			freq:     freq,
			elements: list.New(),
		}
		// Вставляем в отсортированный список частот
		freqNodeElem = c.insertFrequencyNode(freq, freqNode)
		c.freqLists[freq] = freqNodeElem
	}

	// Добавляем элемент в список этой частоты
	freqNode := freqNodeElem.Value.(*FrequencyNode)
	return freqNode.elements.PushBack(item)
}

// insertFrequencyNode вставляет FrequencyNode в отсортированный список
func (c *LFUCache) insertFrequencyNode(freq int, freqNode *FrequencyNode) *list.Element {
	// Если список пуст или частота меньше первого элемента
	if c.freqNodes.Len() == 0 || freq < c.freqNodes.Front().Value.(*FrequencyNode).freq {
		return c.freqNodes.PushFront(freqNode)
	}

	// Ищем позицию для вставки
	for e := c.freqNodes.Front(); e != nil; e = e.Next() {
		currentFreq := e.Value.(*FrequencyNode).freq
		if freq == currentFreq {
			// Уже существует
			return e
		}
		if freq < currentFreq {
			return c.freqNodes.InsertBefore(freqNode, e)
		}
	}

	// Если больше всех, вставляем в конец
	return c.freqNodes.PushBack(freqNode)
}

// removeFromFrequencyList удаляет элемент из списка частот
func (c *LFUCache) removeFromFrequencyList(freq int, elem *list.Element) {
	if freqNodeElem, ok := c.freqLists[freq]; ok {
		freqNode := freqNodeElem.Value.(*FrequencyNode)
		freqNode.elements.Remove(elem)

		// Если список элементов пуст, удаляем FrequencyNode
		if freqNode.elements.Len() == 0 {
			c.freqNodes.Remove(freqNodeElem)
			delete(c.freqLists, freq)
		}
	}
}

// getFrequencyList получает список элементов для заданной частоты
func (c *LFUCache) getFrequencyList(freq int) *list.List {
	if elem, ok := c.freqLists[freq]; ok {
		return elem.Value.(*FrequencyNode).elements
	}
	return nil
}

// evict удаляет наименее часто используемый элемент
func (c *LFUCache) evict() {
	if c.freqNodes.Len() == 0 {
		return
	}

	// Берем первый FrequencyNode (с минимальной частотой)
	minFreqNodeElem := c.freqNodes.Front()
	minFreqNode := minFreqNodeElem.Value.(*FrequencyNode)

	// Удаляем первый элемент из списка (LRU в пределах одной частоты)
	if minFreqNode.elements.Len() > 0 {
		lruElem := minFreqNode.elements.Front()
		item := lruElem.Value.(*CacheItem)

		// Удаляем из всех структур
		minFreqNode.elements.Remove(lruElem)
		delete(c.items, item.key)

		// Если список частот пуст, удаляем FrequencyNode
		if minFreqNode.elements.Len() == 0 {
			c.freqNodes.Remove(minFreqNodeElem)
			delete(c.freqLists, minFreqNode.freq)
		}
	}
}

// Size возвращает текущий размер кэша
func (c *LFUCache) Size() int {
	return len(c.items)
}

// Clear очищает кэш
func (c *LFUCache) Clear() {
	c.items = make(map[interface{}]*list.Element)
	c.freqLists = make(map[int]*list.Element)
	c.freqNodes.Init()
	c.minFreq = 0
}
