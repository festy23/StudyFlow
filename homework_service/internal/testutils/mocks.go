package testutils

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) Send(ctx context.Context, topic string, message interface{}) error {
	args := m.Called(ctx, topic, message)
	return args.Error(0)
}

func (m *MockKafkaProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}
