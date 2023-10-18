package cache

import (
	"github.com/stretchr/testify/mock"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Add(key []byte) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCache) Get(key []byte) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}
