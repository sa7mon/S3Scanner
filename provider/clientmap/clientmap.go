package clientmap

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"sync"
)

type ClientMap struct {
	sync.Mutex
	inner map[string]*s3.Client
}

func New() *ClientMap {
	return &ClientMap{
		Mutex: sync.Mutex{},
		inner: make(map[string]*s3.Client),
	}
}

func WithCapacity(cap int) *ClientMap {
	return &ClientMap{
		Mutex: sync.Mutex{},
		inner: make(map[string]*s3.Client, cap),
	}
}

func (m *ClientMap) Get(key string) *s3.Client {
	m.Lock()
	defer m.Unlock()
	if v, ok := m.inner[key]; ok {
		return v
	}
	return nil
}

func (m *ClientMap) Set(key string, value *s3.Client) {
	m.Lock()
	m.inner[key] = value
	m.Unlock()
}

func (m *ClientMap) Len() int {
	m.Lock()
	defer m.Unlock()
	return len(m.inner)
}

func (m *ClientMap) Each(fn func(region string, client *s3.Client)) {
	m.Lock()
	for region, client := range m.inner {
		fn(region, client)
	}
	m.Unlock()
}
