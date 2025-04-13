package service

import (
	"context"
	"errors"
	"testing"

	"homework_service/internal/domain"
	"homework_service/internal/repository"
)

type mockAssignmentRepo struct {
	assignments map[string]*domain.Assignment
}

func (m *mockAssignmentRepo) Create(ctx context.Context, assignment *domain.Assignment) error {
	m.assignments[assignment.ID] = assignment
	return nil
}

func (m *mockAssignmentRepo) GetByID(ctx context.Context, id string) (*domain.Assignment, error) {
	assignment, ok := m.assignments[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return assignment, nil
}

func (m *mockAssignmentRepo) Update(ctx context.Context, assignment *domain.Assignment) error {
	if _, ok := m.assignments[assignment.ID]; !ok {
		return repository.ErrNotFound
	}
	m.assignments[assignment.ID] = assignment
	return nil
}

func (m *mockAssignmentRepo) Delete(ctx context.Context, id string) error {
	if _, ok := m.assignments[id]; !ok {
		return repository.ErrNotFound
	}
	delete(m.assignments, id)
	return nil
}

func (m *mockAssignmentRepo) ListByTutorID(ctx context.Context, tutorID string) ([]*domain.Assignment, error) {
	var result []*domain.Assignment
	for _, a := range m.assignments {
		if a.TutorID == tutorID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAssignmentRepo) ListByStudentID(ctx context.Context, studentID string) ([]*domain.Assignment, error) {
	var result []*domain.Assignment
	for _, a := range m.assignments {
		if a.StudentID == studentID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAssignmentRepo) Close() error {
	return nil
}

type mockUserClient struct {
	users     map[string]bool
	pairs     map[string]map[string]bool
	userRoles map[string]string
}

func (m *mockUserClient) UserExists(ctx context.Context, userID string) bool {
	return m.users[userID]
}

func (m *mockUserClient) IsPair(ctx context.Context, tutorID, studentID string) bool {
	if pairs, ok := m.pairs[tutorID]; ok {
		return pairs[studentID]
	}
	return false
}

type mockFileClient struct {
	files map[string]bool
}

func (m *mockFileClient) FileExists(ctx context.Context, fileID string) bool {
	return m.files[fileID]
}

func TestAssignmentService_CreateAssignment(t *testing.T) {
	repo := &mockAssignmentRepo{assignments: make(map[string]*domain.Assignment)}
	userClient := &mockUserClient{
		users: map[string]bool{"tutor1": true, "student1": true},
		pairs: map[string]map[string]bool{"tutor1": {"student1": true}},
	}
	fileClient := &mockFileClient{files: map[string]bool{"file1": true}}

	service := NewAssignmentService(repo, userClient, fileClient)

	tests := []struct {
		name        string
		ctx         context.Context
		assignment  *domain.Assignment
		wantErr     bool
		errContains string
	}{
		{
			name: "successful creation",
			ctx:  context.WithValue(context.Background(), "user_role", "tutor"),
			assignment: &domain.Assignment{
				TutorID:     "tutor1",
				StudentID:   "student1",
				Title:       "Test",
				Description: "Description",
				FileID:      "file1",
			},
			wantErr: false,
		},
		{
			name: "missing required fields",
			ctx:  context.WithValue(context.Background(), "user_role", "tutor"),
			assignment: &domain.Assignment{
				TutorID:   "tutor1",
				StudentID: "student1",
			},
			wantErr:     true,
			errContains: "invalid arguments",
		},
		{
			name: "not tutor role",
			ctx:  context.WithValue(context.Background(), "user_role", "student"),
			assignment: &domain.Assignment{
				TutorID:     "tutor1",
				StudentID:   "student1",
				Title:       "Test",
				Description: "Description",
			},
			wantErr:     true,
			errContains: "permission denied",
		},
		{
			name: "student not found",
			ctx:  context.WithValue(context.Background(), "user_role", "tutor"),
			assignment: &domain.Assignment{
				TutorID:     "tutor1",
				StudentID:   "nonexistent",
				Title:       "Test",
				Description: "Description",
			},
			wantErr:     true,
			errContains: "student not found",
		},
		{
			name: "not a tutor-student pair",
			ctx:  context.WithValue(context.Background(), "user_role", "tutor"),
			assignment: &domain.Assignment{
				TutorID:     "tutor1",
				StudentID:   "student2",
				Title:       "Test",
				Description: "Description",
			},
			wantErr:     true,
			errContains: "not a tutor-student pair",
		},
		{
			name: "file not found",
			ctx:  context.WithValue(context.Background(), "user_role", "tutor"),
			assignment: &domain.Assignment{
				TutorID:     "tutor1",
				StudentID:   "student1",
				Title:       "Test",
				Description: "Description",
				FileID:      "nonexistent",
			},
			wantErr:     true,
			errContains: "file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateAssignment(tt.ctx, tt.assignment)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAssignment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !errors.Is(err, errors.New(tt.errContains)) && err.Error() != tt.errContains {
					t.Errorf("CreateAssignment() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestAssignmentService_GetAssignment(t *testing.T) {
	repo := &mockAssignmentRepo{assignments: map[string]*domain.Assignment{
		"1": {
			ID:        "1",
			TutorID:   "tutor1",
			StudentID: "student1",
		},
	}}
	service := NewAssignmentService(repo, nil, nil)

	tests := []struct {
		name        string
		ctx         context.Context
		id          string
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful get by tutor",
			ctx:     context.WithValue(context.Background(), "user_id", "tutor1"),
			id:      "1",
			wantErr: false,
		},
		{
			name:    "successful get by student",
			ctx:     context.WithValue(context.Background(), "user_id", "student1"),
			id:      "1",
			wantErr: false,
		},
		{
			name:        "not found",
			ctx:         context.WithValue(context.Background(), "user_id", "tutor1"),
			id:          "nonexistent",
			wantErr:     true,
			errContains: repository.ErrNotFound.Error(),
		},
		{
			name:        "permission denied",
			ctx:         context.WithValue(context.Background(), "user_id", "other"),
			id:          "1",
			wantErr:     true,
			errContains: ErrPermissionDenied.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetAssignment(tt.ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAssignment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !errors.Is(err, errors.New(tt.errContains)) && err.Error() != tt.errContains {
					t.Errorf("GetAssignment() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestAssignmentService_UpdateAssignment(t *testing.T) {
	repo := &mockAssignmentRepo{assignments: map[string]*domain.Assignment{
		"1": {
			ID:        "1",
			TutorID:   "tutor1",
			StudentID: "student1",
		},
	}}
	service := NewAssignmentService(repo, nil, nil)

	tests := []struct {
		name        string
		ctx         context.Context
		assignment  *domain.Assignment
		wantErr     bool
		errContains string
	}{
		{
			name: "successful update",
			ctx:  context.WithValue(context.Background(), "user_id", "tutor1"),
			assignment: &domain.Assignment{
				ID:        "1",
				TutorID:   "tutor1",
				StudentID: "student1",
			},
			wantErr: false,
		},
		{
			name: "not found",
			ctx:  context.WithValue(context.Background(), "user_id", "tutor1"),
			assignment: &domain.Assignment{
				ID:        "nonexistent",
				TutorID:   "tutor1",
				StudentID: "student1",
			},
			wantErr:     true,
			errContains: repository.ErrNotFound.Error(),
		},
		{
			name: "permission denied",
			ctx:  context.WithValue(context.Background(), "user_id", "other"),
			assignment: &domain.Assignment{
				ID:        "1",
				TutorID:   "tutor1",
				StudentID: "student1",
			},
			wantErr:     true,
			errContains: ErrPermissionDenied.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateAssignment(tt.ctx, tt.assignment)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateAssignment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !errors.Is(err, errors.New(tt.errContains)) && err.Error() != tt.errContains {
					t.Errorf("UpdateAssignment() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestAssignmentService_DeleteAssignment(t *testing.T) {
	repo := &mockAssignmentRepo{assignments: map[string]*domain.Assignment{
		"1": {
			ID:        "1",
			TutorID:   "tutor1",
			StudentID: "student1",
		},
	}}
	service := NewAssignmentService(repo, nil, nil)

	tests := []struct {
		name        string
		ctx         context.Context
		id          string
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful delete",
			ctx:     context.WithValue(context.Background(), "user_id", "tutor1"),
			id:      "1",
			wantErr: false,
		},
		{
			name:        "not found",
			ctx:         context.WithValue(context.Background(), "user_id", "tutor1"),
			id:          "nonexistent",
			wantErr:     true,
			errContains: repository.ErrNotFound.Error(),
		},
		{
			name:        "permission denied",
			ctx:         context.WithValue(context.Background(), "user_id", "other"),
			id:          "1",
			wantErr:     true,
			errContains: ErrPermissionDenied.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteAssignment(tt.ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteAssignment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !errors.Is(err, errors.New(tt.errContains)) && err.Error() != tt.errContains {
					t.Errorf("DeleteAssignment() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestAssignmentService_ListAssignmentsByTutor(t *testing.T) {
	repo := &mockAssignmentRepo{assignments: map[string]*domain.Assignment{
		"1": {
			ID:        "1",
			TutorID:   "tutor1",
			StudentID: "student1",
		},
		"2": {
			ID:        "2",
			TutorID:   "tutor2",
			StudentID: "student1",
		},
	}}
	service := NewAssignmentService(repo, nil, nil)

	tests := []struct {
		name        string
		ctx         context.Context
		tutorID     string
		wantCount   int
		wantErr     bool
		errContains string
	}{
		{
			name:      "successful list",
			ctx:       context.WithValue(context.Background(), "user_id", "tutor1"),
			tutorID:   "tutor1",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:        "permission denied",
			ctx:         context.WithValue(context.Background(), "user_id", "other"),
			tutorID:     "tutor1",
			wantErr:     true,
			errContains: ErrPermissionDenied.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignments, err := service.ListAssignmentsByTutor(tt.ctx, tt.tutorID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAssignmentsByTutor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !errors.Is(err, errors.New(tt.errContains)) && err.Error() != tt.errContains {
					t.Errorf("ListAssignmentsByTutor() error = %v, should contain %v", err, tt.errContains)
				}
			}
			if !tt.wantErr && len(assignments) != tt.wantCount {
				t.Errorf("ListAssignmentsByTutor() count = %v, want %v", len(assignments), tt.wantCount)
			}
		})
	}
}

func TestAssignmentService_ListAssignmentsByStudent(t *testing.T) {
	repo := &mockAssignmentRepo{assignments: map[string]*domain.Assignment{
		"1": {
			ID:        "1",
			TutorID:   "tutor1",
			StudentID: "student1",
		},
		"2": {
			ID:        "2",
			TutorID:   "tutor1",
			StudentID: "student2",
		},
	}}
	service := NewAssignmentService(repo, nil, nil)

	tests := []struct {
		name        string
		ctx         context.Context
		studentID   string
		wantCount   int
		wantErr     bool
		errContains string
	}{
		{
			name:      "successful list",
			ctx:       context.WithValue(context.Background(), "user_id", "student1"),
			studentID: "student1",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:        "permission denied",
			ctx:         context.WithValue(context.Background(), "user_id", "other"),
			studentID:   "student1",
			wantErr:     true,
			errContains: ErrPermissionDenied.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignments, err := service.ListAssignmentsByStudent(tt.ctx, tt.studentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAssignmentsByStudent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !errors.Is(err, errors.New(tt.errContains)) && err.Error() != tt.errContains {
					t.Errorf("ListAssignmentsByStudent() error = %v, should contain %v", err, tt.errContains)
				}
			}
			if !tt.wantErr && len(assignments) != tt.wantCount {
				t.Errorf("ListAssignmentsByStudent() count = %v, want %v", len(assignments), tt.wantCount)
			}
		})
	}
}
