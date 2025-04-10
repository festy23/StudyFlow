package grpc_test

import (
	"context"
	"homework_service/internal/domain"
	"homework_service/internal/service/mocks"
	"homework_service/pkg/api"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHomeworkHandler(t *testing.T) {
	t.Run("CreateAssignment - success", func(t *testing.T) {
		assignmentService := new(mocks.AssignmentService)
		assignmentService.On("CreateAssignment", mock.Anything, mock.Anything).
			Return(&domain.Assignment{
				ID:        "1",
				TutorID:   "tutor1",
				StudentID: "student1",
				Title:     "title",
			}, nil)

		handler := NewHomeworkHandler(assignmentService, nil, nil)

		ctx := context.WithValue(context.Background(), "user_id", "tutor1")
		resp, err := handler.CreateAssignment(ctx, &api.CreateAssignmentRequest{
			TutorId:   "tutor1",
			StudentId: "student1",
			Title:     "title",
		})

		assert.NoError(t, err)
		assert.Equal(t, "1", resp.Id)
	})
}
