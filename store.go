package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type record struct {
	Key, URL string
}

type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
	file *os.File
}

func (s *URLStore) close() {
	err := s.file.Close()
	fmt.Printf("file closed err = %s", err)
}

func NewURLStore(filename string) *URLStore {
	store := &URLStore{
		urls: make(map[string]string),
	}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("URLStore: ", err)
	}

	store.file = f

	if err := store.load(); err != nil {
		log.Println("Error loading data in URLStore: ", err)
	}

	return store
}

func (s *URLStore) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.urls[key]
}

func (s *URLStore) Set(key, url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, present := s.urls[key]; present {
		return false
	}
	s.urls[key] = url
	return true
}

func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

func (s *URLStore) contains(url string) *string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.urls {
		if v == url {
			return &k
		}
	}
	return nil
}

func (s *URLStore) Put(url string) string {

	if key := s.contains(url); key != nil {
		return *key
	}

	for {
		key := genKey(s.Count()) // generate the short URL
		if ok := s.Set(key, url); ok {
			if err := store.save(key, url); err != nil {
				log.Println("Error saving to URLStore: ", err)
			}
			return key
		}
	}
}

func (s *URLStore) save(key, url string) error {
	e := gob.NewEncoder(s.file)
	return e.Encode(record{key, url})
}

func (s *URLStore) load() error {
	if _, err := s.file.Seek(0, 0); err != nil {
		return err
	}

	d := gob.NewDecoder(s.file)
	var err error
	for err != io.EOF {
		var r record
		if err = d.Decode(&r); err == nil {
			s.Set(r.Key, r.URL)
		}
	}

	if err == io.EOF {
		return nil
	}

	return err
}
