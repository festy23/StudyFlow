package service

import (
	"schedule_service/internal/database/repo"
	pb "schedule_service/pkg/api"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertrepoLessonToProto(lesson *repo.Lesson) *pb.Lesson {
	protoLesson := &pb.Lesson{
		Id:        lesson.ID,
		SlotId:    lesson.SlotID,
		StudentId: lesson.StudentID,
		Status:    lesson.Status,
		IsPaid:    lesson.IsPaid,
		CreatedAt: timestamppb.New(lesson.CreatedAt),
		EditedAt:  timestamppb.New(lesson.EditedAt),
	}

	if lesson.ConnectionLink != nil {
		protoLesson.ConnectionLink = lesson.ConnectionLink
	}

	if lesson.PriceRub != nil {
		protoLesson.PriceRub = lesson.PriceRub
	}

	if lesson.PaymentInfo != nil {
		protoLesson.PaymentInfo = lesson.PaymentInfo
	}

	return protoLesson
}

func createListLessonsResponse(lessons []repo.Lesson) *pb.ListLessonsResponse {
	protoLessons := make([]*pb.Lesson, 0, len(lessons))

	for _, lesson := range lessons {
		lessonCopy := lesson
		protoLesson := convertrepoLessonToProto(&lessonCopy)
		protoLessons = append(protoLessons, protoLesson)
	}

	return &pb.ListLessonsResponse{
		Lessons: protoLessons,
	}
}

// import (
// 	"errors"
// 	pb "schedule_service/pkg/api"
// 	"strings"
// )

// func filterLessons(lessons []*pb.Lesson, statusFilters []pb.LessonStatusFilter) []*pb.Lesson {
// 	if len(statusFilters) == 0 {
// 		return lessons
// 	}

// 	filterSet := make(map[pb.LessonStatusFilter]struct{})
// 	for _, f := range statusFilters {
// 		filterSet[f] = struct{}{}
// 	}

// 	result := make([]*pb.Lesson, 0, len(lessons))
// 	for _, lesson := range lessons {
// 		lessonStatus, err := convertStatusToFilter(lesson.GetStatus())
// 		if err != nil {
// 			continue
// 		}

// 		if _, ok := filterSet[lessonStatus]; ok {
// 			result = append(result, lesson)
// 		}
// 	}
// 	return result

// }

// func convertStatusToFilter(status string) (pb.LessonStatusFilter, error) {
// 	switch strings.ToUpper(status) {
// 	case "BOOKED":
// 		return pb.LessonStatusFilter_BOOKED, nil
// 	case "CANCELED":
// 		return pb.LessonStatusFilter_CANCELLED, nil
// 	case "COMPLETED":
// 		return pb.LessonStatusFilter_COMPLETED, nil
// 	}

// 	return pb.LessonStatusFilter_BOOKED, errors.New("Unknown status")

// }
