package clientmap

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"sync"
)

type ClientKey struct {
	Region      string
	Credentials bool
}

type ClientMap struct {
	sync.Mutex
	inner map[ClientKey]*s3.Client
}

func New() *ClientMap {
	return &ClientMap{
		Mutex: sync.Mutex{},
		inner: make(map[ClientKey]*s3.Client),
	}
}

func WithCapacity(cap int) *ClientMap {
	return &ClientMap{
		Mutex: sync.Mutex{},
		inner: make(map[ClientKey]*s3.Client, cap),
	}
}

func (m *ClientMap) Get(region string, credentials bool) *s3.Client {
	m.Lock()
	defer m.Unlock()
	if v, ok := m.inner[ClientKey{Region: region, Credentials: credentials}]; ok {
		return v
	}
	return nil
}

func (m *ClientMap) Set(region string, credentials bool, value *s3.Client) {
	m.Lock()
	m.inner[ClientKey{Region: region, Credentials: credentials}] = value
	m.Unlock()
}

func (m *ClientMap) Len() int {
	m.Lock()
	defer m.Unlock()
	return len(m.inner)
}

func (m *ClientMap) Each(fn func(region string, credentials bool, client *s3.Client)) {
	m.Lock()
	for key, client := range m.inner {
		fn(key.Region, key.Credentials, client)
	}
	m.Unlock()
}
