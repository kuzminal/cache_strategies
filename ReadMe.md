# Cache Strategies Implementation

This project implements two popular caching algorithms in Go: LRU (Least Recently Used) and LFU (Least Frequently Used).

## Project Structure

```
├── cmd/
│   └── app/
│       └── main.go
├── pkg/
│   └── cache/
│       ├── cache.go
│       ├── lru/
│       │   ├── lru_cache.go
│       │   └── lru_cache_test.go
│       └── lfu/
│           ├── lfu_cache.go
│           └── lfu_cache_test.go
├── go.mod
├── go.sum
└── README.md
```

## Implemented Caching Algorithms

### LRU Cache (Least Recently Used)

The LRU cache removes the least recently used items first when the cache reaches its capacity. This implementation uses:

- `container/list` for maintaining the order of usage
- `map[interface{}]*list.Element` for O(1) lookups
- A doubly-linked list to track the access order

Key features:
- O(1) time complexity for Get and Add operations
- Automatic eviction of least recently used items
- Thread-unsafe implementation (intended for single-threaded use)

### LFU Cache (Least Frequently Used)

The LFU cache removes the least frequently used items first. This implementation uses a more complex data structure:

- Multiple frequency lists organized by access count
- A main list of frequency nodes sorted by frequency
- Hash maps for O(1) lookups
- LRU behavior within the same frequency level

Key features:
- O(1) average time complexity for Get and Put operations
- Handles frequency increments efficiently
- Maintains LRU order among items with the same frequency
- Automatic eviction of least frequently used items

## Usage

### Running the Example

The main application in `cmd/app/main.go` demonstrates the LRU cache usage:

```go
func main() {
    cache := lru.NewLRUCache(2)
    log.Println(cache.Add("key1", "value1")) // true
    log.Println(cache.Add("key2", "value2")) // true
    log.Println(cache.Get("key1"))           // "value1", true
    log.Println(cache.Add("key3", "value3")) // true
    log.Println(cache.Get("key2"))           // "", false
}
```

### Using the Caches

```go
// Create LRU cache with capacity 100
lruCache := lru.NewLRUCache(100)

// Create LFU cache with capacity 100
lfuCache := lfu.NewLFUCache(100)

// Basic operations
lruCache.Add("key", "value")
value, exists := lruCache.Get("key")
removed := lruCache.Remove("key")
```

## Dependencies

- Go 1.18+
- `github.com/stretchr/testify` - for assertion testing

## Testing

The project includes unit tests for both cache implementations in their respective directories. Run tests with:

```bash
go test ./...
```

## Design Considerations

1. **Interface-based design**: The cache package defines a common interface for different cache strategies
2. **Memory efficiency**: Uses maps and linked lists for optimal memory usage
3. **Performance**: All core operations are designed to be O(1) time complexity
4. **Extensibility**: Easy to add new cache strategies by implementing the Cache interface

## Potential Improvements

- Add thread-safety with mutexes or RWMutex
- Implement TTL (Time To Live) support
- Add metrics collection (hit rate, miss rate)
- Support for serialization/deserialization
- Context-aware operations for timeout handling
