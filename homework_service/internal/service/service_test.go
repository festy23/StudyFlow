package service_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"homework_service/internal/domain"
	"homework_service/internal/service"
	"homework_service/internal/service/mocks"
)

type mockAssignmentRepo struct {
	assignments map[string]*domain.Assignment
	nextID      int
}

func (m *mockAssignmentRepo) Create(ctx context.Context, assignment *domain.Assignment) error {
	assignment.ID = string(m.nextID)
	m.assignments[assignment.ID] = assignment
	m.nextID++
	return nil
}

func (m *mockAssignmentRepo) GetByID(ctx context.Context, id string) (*domain.Assignment, error) {
	assignment, exists := m.assignments[id]
	if !exists {
		return nil, service.ErrAssignmentNotFound
	}
	return assignment, nil
}

func (m *mockAssignmentRepo) Update(ctx context.Context, assignment *domain.Assignment) error {
	m.assignments[assignment.ID] = assignment
	return nil
}

func (m *mockAssignmentRepo) Delete(ctx context.Context, id string) error {
	delete(m.assignments, id)
	return nil
}

func (m *mockAssignmentRepo) ListByTutorID(ctx context.Context, tutorID string) ([]*domain.Assignment, error) {
	var result []*domain.Assignment
	for _, assignment := range m.assignments {
		if assignment.TutorID == tutorID {
			result = append(result, assignment)
		}
	}
	return result, nil
}

func (m *mockAssignmentRepo) ListByStudentID(ctx context.Context, studentID string) ([]*domain.Assignment, error) {
	var result []*domain.Assignment
	for _, assignment := range m.assignments {
		if assignment.StudentID == studentID {
			result = append(result, assignment)
		}
	}
	return result, nil
}

func TestCreateAssignment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAssignmentService := mocks.NewMockAssignmentService(ctrl)

	assignment := &domain.Assignment{
		ID:    "1",
		Title: "Test Assignment",
	}

	mockAssignmentService.EXPECT().
		CreateAssignment(gomock.Any(), assignment).
		Return(assignment, nil)

	serviceInstance := service.NewAssignmentService(mockAssignmentService)

	result, err := serviceInstance.CreateAssignment(context.Background(), assignment)

	assert.NoError(t, err)
	assert.Equal(t, assignment, result)
}

func TestGetAssignment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAssignmentService := mocks.NewMockAssignmentService(ctrl)

	assignment := &domain.Assignment{
		ID:    "1",
		Title: "Test Assignment",
	}

	mockAssignmentService.EXPECT().
		GetAssignment(gomock.Any(), "1").
		Return(assignment, nil)

	serviceInstance := service.NewAssignmentService(mockAssignmentService)

	result, err := serviceInstance.GetAssignment(context.Background(), "1")

	assert.NoError(t, err)
	assert.Equal(t, assignment, result)
}

func TestCreateFeedback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackService := &mocks.FeedbackService{}
	feedback := &domain.Feedback{ID: "1", Content: "Great job!"}

	mockFeedbackService.On("CreateFeedback", gomock.Any(), feedback).Return(feedback, nil)

	serviceInstance := service.NewFeedbackService(mockFeedbackService)

	result, err := serviceInstance.CreateFeedback(context.Background(), feedback)

	assert.NoError(t, err)
	assert.Equal(t, feedback, result)
}

func TestGetFeedback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackService := &mocks.FeedbackService{}
	feedback := &domain.Feedback{ID: "1", Content: "Great job!"}

	mockFeedbackService.On("GetFeedback", gomock.Any(), "1").Return(feedback, nil)

	serviceInstance := service.NewFeedbackService(mockFeedbackService)

	result, err := serviceInstance.GetFeedback(context.Background(), "1")

	assert.NoError(t, err)
	assert.Equal(t, feedback, result)
}

func TestCreateSubmission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubmissionService := mocks.NewMockSubmissionService(ctrl)

	submission := &domain.Submission{
		ID:      "1",
		Content: "Submission Content",
	}

	mockSubmissionService.EXPECT().
		CreateSubmission(gomock.Any(), submission).
		Return(submission, nil)

	serviceInstance := service.NewSubmissionService(mockSubmissionService)

	result, err := serviceInstance.CreateSubmission(context.Background(), submission)

	assert.NoError(t, err)
	assert.Equal(t, submission, result)
}
