package main

import (
	"encoding/gob"
	"io"
	"log"
	"os"
	"sync"
)

type record struct {
	Key, URL string
}

type URLStore struct {
	urls   map[string]string
	mu     sync.RWMutex
	saveCh chan record
}

func NewURLStore(filename string, saveQueueLength int) *URLStore {
	store := &URLStore{
		urls:   make(map[string]string),
		saveCh: make(chan record, saveQueueLength),
	}

	if err := store.load(filename); err != nil {
		log.Println("Error loading data in URLStore: ", err)
	}

	go store.saveLoop(filename)

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
			s.saveCh <- record{key, url}
			return key
		}
	}
}

func (s *URLStore) saveLoop(filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer file.Close()
	if err != nil {
		log.Fatal("URLStore saveloop: ", err)
		return err
	}

	for {
		r := <-s.saveCh // taking a record from the channel and encoding it
		e := gob.NewEncoder(file)
		if err := e.Encode(r); err != nil {
			log.Println("URLStore:", err)
		}
	}

}

func (s *URLStore) load(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("URLStore: ", err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	d := gob.NewDecoder(file)
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
