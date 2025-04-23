package handler

import (
	"common_library/logging"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
	userpb "userservice/pkg/api"
)

type UserHandler struct {
	c     userpb.UserServiceClient
	cache Cache
}

func NewUserHandler(c userpb.UserServiceClient, cache Cache) *UserHandler {
	return &UserHandler{c: c, cache: cache}
}

func (h *UserHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Group(func(r chi.Router) {
		r.Get("/users/me", h.GetMe)
		r.Get("/users/{id}", h.GetUser)
		r.Patch("/users/{id}", h.UpdateUser)
		r.Get("/tutor-profiles/{id}", h.GetTutorProfile)
		r.Patch("/tutor-profiles/{id}", h.UpdateTutorProfile)
		r.Get("/tutor-students/by-tutor/{id}", h.ListTutorStudentByTutor)
		r.Get("/tutor-students/by-student/{id}", h.ListTutorStudentByStudent)
		r.Get("/tutor-students/{tutor_id}/{student_id}", h.GetTutorStudent)
		r.Patch("/tutor-students/{tutor_id}/{student_id}", h.UpdateTutorStudent)
		r.Delete("/tutor-students/{tutor_id}/{student_id}", h.DeleteTutorStudent)
		r.Post("/tutor-students", h.CreateTutorStudent)
		r.Post("/tutor-students/{tutor_id}/accept", h.AcceptInvitation)
	})
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	handler, err := HandleWithCache[userpb.Empty, userpb.User](
		h.c.GetMe, nil, false,
		h.cache, buildMeKey, 5*time.Minute,
	)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	handler, err := HandleWithCache[userpb.GetUserRequest, userpb.UserPublic](
		h.c.GetUser, getUserParsePath, false,
		h.cache, buildUserPublicKey, 5*time.Minute,
	)
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

	key, err := buildUserKey(r)
	if err == nil {
		h.cache.Delete(r.Context(), key)
	}
	key, err = buildUserPublicKey(r)
	if err == nil {
		h.cache.Delete(r.Context(), key)
	}
	handler(w, r)
}

func (h *UserHandler) GetTutorProfile(w http.ResponseWriter, r *http.Request) {
	handler, err := HandleWithCache[userpb.GetTutorProfileByUserIdRequest, userpb.TutorProfile](
		h.c.GetTutorProfileByUserId, getTutorProfileParsePath, false,
		h.cache, buildTutorProfileKey, 5*time.Minute,
	)
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
	key, err := buildTutorProfileKey(r)
	if err == nil {
		h.cache.Delete(r.Context(), key)
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
	handler, err := HandleWithCache[userpb.GetTutorStudentRequest, userpb.TutorStudent](
		h.c.GetTutorStudent, getTutorStudentParsePath, false,
		h.cache, buildTutorStudentKey, 5*time.Minute,
	)
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

	key, err := buildTutorStudentKey(r)
	if err == nil {
		h.cache.Delete(r.Context(), key)
	}

	handler(w, r)
}

func (h *UserHandler) DeleteTutorStudent(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.DeleteTutorStudentRequest, userpb.Empty](h.c.DeleteTutorStudent, deleteTutorStudentParsePath, false)
	if err != nil {
		panic(err)
	}

	key, err := buildTutorStudentKey(r)
	if err == nil {
		h.cache.Delete(r.Context(), key)
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

func (h *UserHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.AcceptInvitationFromTutorRequest, userpb.Empty](
		h.c.AcceptInvitationFromTutor, acceptInvitationParsePath, false,
	)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func acceptInvitationParsePath(ctx context.Context, httpReq *http.Request, grpcReq *userpb.AcceptInvitationFromTutorRequest) error {
	tutorId := chi.URLParam(httpReq, "tutor_id")
	if tutorId == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "tutorId is required")
	}
	grpcReq.TutorId = tutorId
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Debug(ctx, "user id added to request", zap.Any("req", grpcReq))
	}
	return nil
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

func buildUserKey(r *http.Request) (string, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		return "", fmt.Errorf("missing path param: id")
	}
	return fmt.Sprintf("user:%s", id), nil
}

func buildUserPublicKey(r *http.Request) (string, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		return "", fmt.Errorf("missing path param: id")
	}
	return fmt.Sprintf("user-public:%s", id), nil
}

func buildMeKey(r *http.Request) (string, error) {
	id := r.Header.Get("X-User-Id")
	if id == "" {
		return "", fmt.Errorf("missing header: X-User-Id")
	}
	return fmt.Sprintf("user:%s", id), nil
}

func buildTutorProfileKey(r *http.Request) (string, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		return "", fmt.Errorf("missing path param: id")
	}
	return fmt.Sprintf("tutor-profile:%s", id), nil
}

func buildTutorStudentKey(r *http.Request) (string, error) {
	tutorID := chi.URLParam(r, "tutor_id")
	studentID := chi.URLParam(r, "student_id")
	if tutorID == "" || studentID == "" {
		return "", fmt.Errorf("missing tutor_id or student_id")
	}
	return fmt.Sprintf("tutor-student:%s:%s", tutorID, studentID), nil
}
