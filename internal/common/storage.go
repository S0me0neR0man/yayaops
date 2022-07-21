package common

import (
	"log"
	"reflect"
	"sync"
)

// Getter interface
type Getter interface {
	// Get input is 'key'
	Get(string) (any, bool)
	GetNames() []string
}

// Setter interface
type Setter interface {
	// Set first param is the 'key', all next params (values) must be the same type
	// all values will be added
	Set(string, ...any)
}

// Storage the thread-safe storage
type Storage struct {
	sync.RWMutex
	data map[string]any
}

// NewStorage the constructor
func NewStorage() *Storage {
	s := Storage{}
	s.data = make(map[string]any)
	return &s
}

// Set implementation the Setter
func (s *Storage) Set(key string, values ...any) {
	const pre = "Storage.Set()"
	if len(values) == 0 {
		log.Println(pre, "nothing to set")
		return
	}
	// calc sum of values
	sum := values[0]
	sumReflectValue := reflect.ValueOf(sum).Kind()
	for i := 1; i < len(values); i++ {
		v := reflect.ValueOf(values[i])
		if sumReflectValue != v.Kind() {
			log.Fatal(pre, "values of different types")
		}
		switch sumReflectValue {
		case reflect.Float64:
			sum = sum.(float64) + v.Float()
		case reflect.Int64:
			sum = sum.(int64) + v.Int()
		default:
			log.Fatal(pre, v.Kind(), "not implemented")
		}
	}
	s.Lock()
	s.data[key] = sum
	s.Unlock()
}

// Get implementation the Getter
func (s *Storage) Get(name string) (any, bool) {
	s.RLock()
	v, ok := s.data[name]
	s.RUnlock()
	return v, ok
}

// GetNames implementation the Getter
func (s *Storage) GetNames() []string {
	names := make([]string, len(s.data))
	i := 0
	s.RLock()
	for k := range s.data {
		names[i] = k
		i++
	}
	s.RUnlock()
	return names
}
