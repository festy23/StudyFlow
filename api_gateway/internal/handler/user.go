package handler

import (
	"common_library/logging"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	userpb "userservice/pkg/api"
)

type UserHandler struct {
	c userpb.UserServiceClient
}

func NewUserHandler(c userpb.UserServiceClient) *UserHandler {
	return &UserHandler{c: c}
}

func (h *UserHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Group(func(r chi.Router) {
		r.Get("/users/me", h.GetMe)
		r.Get("/users/{id}", h.GetUser)
		r.Patch("/users/{id}", h.UpdateUser)
		r.Get("/tutor-profiles/{id}", h.GetTutorProfile)
		r.Patch("/tutor-profiles/{id}", h.UpdateTutorProfile)
		r.Get("/users/tutor-students/by-tutor/{id}", h.GetTutorStudent)
		r.Get("/users/tutor-students/by-student/{id}", h.GetTutorStudent)
		r.Get("/users/tutor-profiles/{tutor_id}/{student_id}", h.GetTutorProfile)
		r.Patch("/users/tutor-profiles/{tutor_id}/{student_id}", h.UpdateUser)
		r.Delete("/users/tutor-profiles/{tutor_id}/{student_id}", h.DeleteTutorStudent)
		r.Post("/users/tutor-students", h.CreateTutorStudent)
	})
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.Empty, userpb.User](h.c.GetMe, nil, false)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.GetUserRequest, userpb.UserPublic](h.c.GetUser, getUserParsePath, false)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.UpdateUserRequest, userpb.User](h.c.UpdateUser, updateUserParsePath, true)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *UserHandler) GetTutorProfile(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.GetTutorProfileByUserIdRequest, userpb.TutorProfile](h.c.GetTutorProfileByUserId, getTutorProfileParsePath, false)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *UserHandler) UpdateTutorProfile(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.UpdateTutorProfileRequest, userpb.TutorProfile](h.c.UpdateTutorProfile, updateTutorProfileParsePath, false)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *UserHandler) ListTutorStudentByTutor(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.ListTutorStudentsRequest, userpb.ListTutorStudentsResponse](h.c.ListTutorStudents, listTutorStudentsByTutorParsePath, false)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *UserHandler) ListTutorStudentByStudent(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.ListTutorsForStudentRequest, userpb.ListTutorsForStudentResponse](h.c.ListTutorsForStudent, listTutorStudentsByStudentParsePath, false)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *UserHandler) GetTutorStudent(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.GetTutorStudentRequest, userpb.TutorStudent](h.c.GetTutorStudent, getTutorStudentParsePath, false)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *UserHandler) UpdateTutorStudent(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.UpdateTutorStudentRequest, userpb.TutorStudent](h.c.UpdateTutorStudent, updateTutorStudentParsePath, true)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *UserHandler) DeleteTutorStudent(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.DeleteTutorStudentRequest, userpb.Empty](h.c.DeleteTutorStudent, deleteTutorStudentParsePath, false)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *UserHandler) CreateTutorStudent(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.CreateTutorStudentRequest, userpb.TutorStudent](h.c.CreateTutorStudent, nil, true)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func getUserParsePath(ctx context.Context, httpReq *http.Request, grpcReq *userpb.GetUserRequest) error {
	userId := chi.URLParam(httpReq, "id")
	if userId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "userId is required")
	}
	grpcReq.Id = userId
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Debug(ctx, "user id added to request", zap.Any("req", grpcReq))
	}
	return nil
}

func updateUserParsePath(ctx context.Context, httpReq *http.Request, grpcReq *userpb.UpdateUserRequest) error {
	userId := chi.URLParam(httpReq, "id")
	if userId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "userId is required")
	}
	grpcReq.Id = userId
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Debug(ctx, "user id added to request", zap.Any("req", grpcReq))
	}
	return nil
}

func getTutorProfileParsePath(ctx context.Context, httpReq *http.Request, grpcReq *userpb.GetTutorProfileByUserIdRequest) error {
	userId := chi.URLParam(httpReq, "id")
	if userId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "userId is required")
	}
	grpcReq.UserId = userId
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Debug(ctx, "user id added to request", zap.Any("req", grpcReq))
	}
	return nil
}

func updateTutorProfileParsePath(ctx context.Context, httpReq *http.Request, grpcReq *userpb.UpdateTutorProfileRequest) error {
	userId := chi.URLParam(httpReq, "id")
	if userId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "userId is required")
	}
	grpcReq.UserId = userId
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Debug(ctx, "user id added to request", zap.Any("req", grpcReq))
	}
	return nil
}

func listTutorStudentsByTutorParsePath(ctx context.Context, httpReq *http.Request, grpcReq *userpb.ListTutorStudentsRequest) error {
	userId := chi.URLParam(httpReq, "id")
	if userId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "userId is required")
	}
	grpcReq.TutorId = userId
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Debug(ctx, "user id added to request", zap.Any("req", grpcReq))
	}
	return nil
}

func listTutorStudentsByStudentParsePath(ctx context.Context, httpReq *http.Request, grpcReq *userpb.ListTutorsForStudentRequest) error {
	userId := chi.URLParam(httpReq, "id")
	if userId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "userId is required")
	}
	grpcReq.StudentId = userId
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Debug(ctx, "user id added to request", zap.Any("req", grpcReq))
	}
	return nil
}

func getTutorStudentParsePath(ctx context.Context, httpReq *http.Request, grpcReq *userpb.GetTutorStudentRequest) error {
	tutorId := chi.URLParam(httpReq, "tutor_id")
	if tutorId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "tutorId is required")
	}
	grpcReq.TutorId = tutorId

	studentId := chi.URLParam(httpReq, "student_id")
	if studentId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "studentId is required")
	}
	grpcReq.StudentId = studentId

	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Debug(ctx, "tutor, student ids added to request", zap.Any("req", grpcReq))
	}
	return nil
}

func updateTutorStudentParsePath(ctx context.Context, httpReq *http.Request, grpcReq *userpb.UpdateTutorStudentRequest) error {
	tutorId := chi.URLParam(httpReq, "tutor_id")
	if tutorId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "tutorId is required")
	}
	grpcReq.TutorId = tutorId

	studentId := chi.URLParam(httpReq, "student_id")
	if studentId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "studentId is required")
	}
	grpcReq.StudentId = studentId

	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Debug(ctx, "tutor, student ids added to request", zap.Any("req", grpcReq))
	}
	return nil
}

func deleteTutorStudentParsePath(ctx context.Context, httpReq *http.Request, grpcReq *userpb.DeleteTutorStudentRequest) error {
	tutorId := chi.URLParam(httpReq, "tutor_id")
	if tutorId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "tutorId is required")
	}
	grpcReq.TutorId = tutorId

	studentId := chi.URLParam(httpReq, "student_id")
	if studentId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "studentId is required")
	}
	grpcReq.StudentId = studentId

	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Debug(ctx, "tutor, student ids added to request", zap.Any("req", grpcReq))
	}
	return nil
}
