package main_test

import (
	"context"
	"errors"
	"homework_service/internal/domain"
	"homework_service/internal/repository/mocks"
	"homework_service/internal/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

func TestReminderWorker(t *testing.T) {
	t.Run("process reminders successfully", func(t *testing.T) {
		assignmentRepo := new(mocks.AssignmentRepository)
		kafkaProducer := new(testutils.MockKafkaProducer)
		logger := new(testutils.MockLogger)

		assignments := []*domain.Assignment{
			{
				ID:        "1",
				TutorID:   "tutor1",
				StudentID: "student1",
				Title:     "Test Assignment",
				DueDate:   timePtr(time.Now().Add(12 * time.Hour)),
			},
		}

		assignmentRepo.On("FindAssignmentsDueSoon", mock.Anything, 24*time.Hour).
			Return(assignments, nil)
		kafkaProducer.On("Send", mock.Anything, "assignment-reminders", mock.Anything).
			Return(nil)
		logger.On("Infof", mock.Anything, mock.Anything)

		worker := &ReminderWorker{
			assignmentRepo: assignmentRepo,
			kafkaProducer:  kafkaProducer,
			logger:         logger,
			interval:       100 * time.Millisecond,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go worker.Start(ctx)

		time.Sleep(150 * time.Millisecond)
		assignmentRepo.AssertExpectations(t)
		kafkaProducer.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("handle repository error", func(t *testing.T) {
		assignmentRepo := new(mocks.AssignmentRepository)
		kafkaProducer := new(testutils.MockKafkaProducer)
		logger := new(testutils.MockLogger)

		assignmentRepo.On("FindAssignmentsDueSoon", mock.Anything, 24*time.Hour).
			Return(nil, errors.New("repository error"))
		logger.On("Errorf", mock.Anything, mock.Anything)

		worker := &ReminderWorker{
			assignmentRepo: assignmentRepo,
			kafkaProducer:  kafkaProducer,
			logger:         logger,
			interval:       100 * time.Millisecond,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go worker.Start(ctx)
		time.Sleep(150 * time.Millisecond)

		assignmentRepo.AssertExpectations(t)
		logger.AssertExpectations(t)
	})
}

func timePtr(t time.Time) *time.Time {
	return &t
}
