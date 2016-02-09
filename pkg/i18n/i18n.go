package i18n

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
)

var (
	mu      sync.RWMutex // syncrhonizes everything in this package.
	bundles map[string]Bundle
)

func init() {
	mu.Lock()
	bundles = make(map[string]Bundle, 0)
	mu.Unlock()
}

type Bundle struct {
	Name  string
	items map[string]Item
}

func (b *Bundle) T(key string) string {
	var result string
	mu.RLock()
	item, ok := b.items[key]
	mu.RUnlock()
	if ok {
		result = item.Message
	}
	if result == "" {
		result = fmt.Sprintf("[%s]", key)
	}
	return result
}

func (b *Bundle) Has(key string) bool {
	mu.RLock()
	_, ok := b.items[key]
	mu.RUnlock()
	return ok
}

type Item struct {
	ID          string `json:"-"` // Translation key
	Message     string `json:"message"`
	Description string `json:"description"`
}

func AddBundle(name string, data []byte) error {
	var items map[string]Item
	dec := json.NewDecoder(strings.NewReader(string(data)))
	for {
		if err := dec.Decode(&items); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}

	for k, v := range items {
		v.ID = k
		items[k] = v
	}

	b := Bundle{
		Name:  name,
		items: items,
	}
	mu.Lock()
	bundles[name] = b
	mu.Unlock()
	return nil
}

func Tfunc(bundle string) (func(key string) string, error) {
	mu.RLock()
	b, ok := bundles[bundle]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("bundle not found")
	}
	return b.T, nil
}

func Hasfunc(bundle string) (func(key string) bool, error) {
	mu.RLock()
	b, ok := bundles[bundle]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("bundle not found")
	}
	return b.Has, nil
}
