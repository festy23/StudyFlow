package handler

import (
	"common_library/logging"
	"context"
	"fmt"
	"net/http"
	schedulepb "schedule_service/pkg/api"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ScheduleHandler struct {
	c schedulepb.ScheduleServiceClient
}

func NewScheduleHandler(c schedulepb.ScheduleServiceClient) *ScheduleHandler {
	return &ScheduleHandler{c: c}
}

func (h *ScheduleHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Group(func(r chi.Router) {
		r.Post("/slots", h.CreateSlot)
		r.Get("/slots/{id}", h.GetSlot)
		r.Patch("/slots/{id}", h.UpdateSlot)
		r.Delete("/slots/{id}", h.DeleteSlot)
		r.Get("/slots/by-tutor/{tutor_id}", h.ListSlotsByTutor)

		r.Get("/lessons", h.ListLessons)
		r.Post("/lessons", h.CreateLesson)
		r.Get("/lessons/{id}", h.GetLesson)
		r.Patch("/lessons/{id}", h.UpdateLesson)
		r.Post("/lessons/{id}/cancel", h.CancelLesson)
	})
}

func parseIDParam(r *http.Request, name string) (string, error) {
	id := chi.URLParam(r, name)
	if id == "" {
		return "", fmt.Errorf("missing required path param: %s", name)
	}
	return id, nil
}

func parseListSlotsByTutor(ctx context.Context, r *http.Request, req *schedulepb.ListSlotsByTutorRequest) error {
	tutorID, err := parseIDParam(r, "tutor_id")
	if err != nil {
		return err
	}
	req.TutorId = tutorID
	if only := r.URL.Query().Get("only_available"); only == "true" {
		v := true
		req.OnlyAvailable = &v
	}
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Debug(ctx, "parsed listSlotsByTutor", zap.Any("req", req))
	}
	return nil
}

func parseGetSlot(ctx context.Context, r *http.Request, req *schedulepb.GetSlotRequest) error {
	id, err := parseIDParam(r, "id")
	if err != nil {
		return err
	}
	req.Id = id
	return nil
}

func parseDeleteSlot(ctx context.Context, r *http.Request, req *schedulepb.DeleteSlotRequest) error {
	id, err := parseIDParam(r, "id")
	if err != nil {
		return err
	}
	req.Id = id
	return nil
}

func parseUpdateSlot(ctx context.Context, r *http.Request, req *schedulepb.UpdateSlotRequest) error {
	id, err := parseIDParam(r, "id")
	if err != nil {
		return err
	}
	req.Id = id
	return nil
}

func parseGetLesson(ctx context.Context, r *http.Request, req *schedulepb.GetLessonRequest) error {
	id, err := parseIDParam(r, "id")
	if err != nil {
		return err
	}
	req.Id = id
	return nil
}

func parseUpdateLesson(ctx context.Context, r *http.Request, req *schedulepb.UpdateLessonRequest) error {
	id, err := parseIDParam(r, "id")
	if err != nil {
		return err
	}
	req.Id = id
	return nil
}

func parseCancelLesson(ctx context.Context, r *http.Request, req *schedulepb.CancelLessonRequest) error {
	id, err := parseIDParam(r, "id")
	if err != nil {
		return err
	}
	req.Id = id
	return nil
}

func parseListLessons(ctx context.Context, r *http.Request) (context.Context, any, error) {
	q := r.URL.Query()
	tutorID := q.Get("tutor_id")
	studentID := q.Get("student_id")
	statusParams := q["status_filter"]

	switch {
	case tutorID != "" && studentID != "":
		req := &schedulepb.ListLessonsByPairRequest{TutorId: tutorID, StudentId: studentID}
		for _, s := range statusParams {
			req.StatusFilter = append(req.StatusFilter, parseStatus(s))
		}
		return ctx, req, nil
	case tutorID != "":
		req := &schedulepb.ListLessonsByTutorRequest{TutorId: tutorID}
		for _, s := range statusParams {
			req.StatusFilter = append(req.StatusFilter, parseStatus(s))
		}
		return ctx, req, nil
	case studentID != "":
		req := &schedulepb.ListLessonsByStudentRequest{StudentId: studentID}
		for _, s := range statusParams {
			req.StatusFilter = append(req.StatusFilter, parseStatus(s))
		}
		return ctx, req, nil
	default:
		return nil, nil, fmt.Errorf("invalid combination of parameters")
	}
}

func parseStatus(s string) schedulepb.LessonStatusFilter {
	switch strings.ToUpper(s) {
	case "BOOKED":
		return schedulepb.LessonStatusFilter_BOOKED
	case "CANCELLED":
		return schedulepb.LessonStatusFilter_CANCELLED
	case "COMPLETED":
		return schedulepb.LessonStatusFilter_COMPLETED
	default:
		return schedulepb.LessonStatusFilter_BOOKED
	}
}

func (h *ScheduleHandler) CreateSlot(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[schedulepb.CreateSlotRequest, schedulepb.Slot](h.c.CreateSlot, nil, true)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *ScheduleHandler) GetSlot(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[schedulepb.GetSlotRequest, schedulepb.Slot](h.c.GetSlot, parseGetSlot, false)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *ScheduleHandler) UpdateSlot(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[schedulepb.UpdateSlotRequest, schedulepb.Slot](h.c.UpdateSlot, parseUpdateSlot, true)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *ScheduleHandler) DeleteSlot(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[schedulepb.DeleteSlotRequest, schedulepb.Empty](h.c.DeleteSlot, parseDeleteSlot, false)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *ScheduleHandler) ListSlotsByTutor(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[schedulepb.ListSlotsByTutorRequest, schedulepb.ListSlotsResponse](h.c.ListSlotsByTutor, parseListSlotsByTutor, false)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *ScheduleHandler) GetLesson(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[schedulepb.GetLessonRequest, schedulepb.Lesson](h.c.GetLesson, parseGetLesson, false)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *ScheduleHandler) CreateLesson(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[schedulepb.CreateLessonRequest, schedulepb.Lesson](h.c.CreateLesson, nil, true)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *ScheduleHandler) UpdateLesson(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[schedulepb.UpdateLessonRequest, schedulepb.Lesson](h.c.UpdateLesson, parseUpdateLesson, true)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *ScheduleHandler) CancelLesson(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[schedulepb.CancelLessonRequest, schedulepb.Lesson](h.c.CancelLesson, parseCancelLesson, false)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *ScheduleHandler) ListLessons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, customReq, err := parseListLessons(ctx, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch req := customReq.(type) {
	case *schedulepb.ListLessonsByTutorRequest:
		handler, err := Handle[schedulepb.ListLessonsByTutorRequest, schedulepb.ListLessonsResponse](
			h.c.ListLessonsByTutor,
			func(_ context.Context, _ *http.Request, grpcReq *schedulepb.ListLessonsByTutorRequest) error {
				grpcReq.TutorId = req.TutorId
				grpcReq.StatusFilter = req.StatusFilter
				return nil
			}, false,
		)
		if err != nil {
			panic(err)
		}
		handler(w, r.WithContext(context.WithValue(ctx, "req", req)))
	case *schedulepb.ListLessonsByStudentRequest:
		handler, err := Handle[schedulepb.ListLessonsByStudentRequest, schedulepb.ListLessonsResponse](
			h.c.ListLessonsByStudent,
			func(_ context.Context, _ *http.Request, grpcReq *schedulepb.ListLessonsByStudentRequest) error {
				grpcReq.StudentId = req.StudentId
				grpcReq.StatusFilter = req.StatusFilter
				return nil
			}, false)
		if err != nil {
			panic(err)
		}
		handler(w, r.WithContext(context.WithValue(ctx, "req", req)))
	case *schedulepb.ListLessonsByPairRequest:
		handler, err := Handle[schedulepb.ListLessonsByPairRequest, schedulepb.ListLessonsResponse](
			h.c.ListLessonsByPair,
			func(_ context.Context, _ *http.Request, grpcReq *schedulepb.ListLessonsByPairRequest) error {
				grpcReq.TutorId = req.TutorId
				grpcReq.StudentId = req.StudentId
				grpcReq.StatusFilter = req.StatusFilter
				return nil
			}, false)
		if err != nil {
			panic(err)
		}
		handler(w, r.WithContext(context.WithValue(ctx, "req", req)))
	default:
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
	}
}
