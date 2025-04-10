package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"homework_service/internal/domain"
	"homework_service/internal/service"
	"homework_service/internal/service/mocks"
)

func TestCreateAssignment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := new(mocks.MockAssignmentRepository)
		userClient := new(mocks.MockUserClient)
		fileClient := new(mocks.MockFileClient)

		userClient.On("UserExists", mock.Anything, "student1").Return(true)
		userClient.On("IsPair", mock.Anything, "tutor1", "student1").Return(true)
		fileClient.On("FileExists", mock.Anything, "file1").Return(true)
		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		svc := service.NewAssignmentService(repo, userClient, fileClient)

		assignment := &domain.Assignment{
			TutorID:     "tutor1",
			StudentID:   "student1",
			Title:       "Test",
			Description: "Desc",
			FileID:      "file1",
		}

		ctx := context.WithValue(context.Background(), "user_role", "tutor")
		result, err := svc.CreateAssignment(ctx, assignment)

		require.NoError(t, err)
		require.NotNil(t, result)
		repo.AssertExpectations(t)
		userClient.AssertExpectations(t)
		fileClient.AssertExpectations(t)
	})
}
