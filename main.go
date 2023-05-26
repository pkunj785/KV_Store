package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

type Storer[K comparable, V any] interface {
	Put(K, V) error
	Get(K) (V, error)
	Update(K, V) error
	Delete(K) error
}

///////////////////////////////

type KVStore[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

//////////////////////////////////////

func NewKVStore[K comparable, V any]() *KVStore[K, V] {
	return &KVStore[K, V]{
		data: make(map[K]V),
	}
}

////////////////////////////////////

func (s *KVStore[K, V]) Put(key K, val V) error {

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val

	return nil
}

////////////////////////////////////////////////

func (s *KVStore[K, V]) Get(key K) (V, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.data[key]

	if !ok {
		return val, fmt.Errorf("the key (%v) does not exist", key)
	}

	return val, nil
}

/////////////////////////////////////////////////

func (s *KVStore[K, V]) Update(key K, val V) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.Has(key) {
		return fmt.Errorf("the key (%v) does not exist", key)
	}

	s.data[key] = val

	return nil
}

/////////////////////////////////////////////////////

func (s *KVStore[K, V]) Delete(key K) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[key]
	if !ok {
		return fmt.Errorf("the key (%v) does not exits", key)
	}

	delete(s.data, key)

	fmt.Printf("the key (%v) successfully deleted\n", key)
	return nil
}

////////////////////////////////////////////////////////

type Server struct {
	ListenAddr string
	Storage    Storer[string, string]
}

func NewServer(listenAddr string) *Server {
	return &Server{
		ListenAddr: listenAddr,
		Storage:    NewKVStore[string, string](),
	}
}

////////////////////////################################################
// handler

func (s *Server) handlePut(c echo.Context) error {
	key := c.Param("key")
	value := c.Param("value")

	s.Storage.Put(key, value)

	return c.JSON(http.StatusOK, map[string]string{"msg": "ok"})
}

func (s *Server) handleGet(c echo.Context) error {
	key := c.Param("key")

	value, err := s.Storage.Get(key)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{"value": "key not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"value": value})
}

func (s *Server) handleUpdate(c echo.Context) error {
	key := c.Param("key")
	value := c.Param("value")

	if err := s.Storage.Update(key, value); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"msg": "updated"})
}

func (s *Server) handleDelete(c echo.Context) error {
	key := c.Param("key")

	if err := s.Storage.Delete(key); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"msg": "deleted"})
}

///////////////////#########################################

func (s *Server) Start() {
	fmt.Printf("HTTP server is running on port %s", s.ListenAddr)

	e := echo.New()

	// handle here
	e.GET("/put/:key/:value", s.handlePut)
	e.GET("get/:key", s.handleGet)
	e.GET("/update/:key/:value", s.handleUpdate)
	e.GET("delete/:key", s.handleDelete)

	e.Start(s.ListenAddr)
}

func main() {
	s := NewServer(":3000")

	s.Start()

}

////////////////////////////////////////////

func (s *KVStore[K, V]) Has(key K) bool {
	_, ok := s.data[key]

	return ok
}
