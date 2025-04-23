package service_test

import (
	"context"
	"errors"
	api2 "fileservice/pkg/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	errdefs "paymentservice/internal/errors"
	"paymentservice/internal/mocks"
	"paymentservice/internal/models"
	"paymentservice/internal/service"
	"schedule_service/pkg/api"
	"testing"
	"time"
)

func setup(t *testing.T) (*gomock.Controller, *service.PaymentService, *mocks.MockIPaymentRepo, *mocks.MockUserServiceClient, *mocks.MockFileServiceClient, *mocks.MockScheduleServiceClient) {
	ctrl := gomock.NewController(t)

	mockRepo := mocks.NewMockIPaymentRepo(ctrl)
	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	mockFileClient := mocks.NewMockFileServiceClient(ctrl)
	mockScheduleClient := mocks.NewMockScheduleServiceClient(ctrl)

	svc := service.NewPaymentService(mockRepo, mockUserClient, mockFileClient, mockScheduleClient)
	return ctrl, svc, mockRepo, mockUserClient, mockFileClient, mockScheduleClient
}
func TestSubmitPaymentReceipt(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl, svc, mockRepo, _, _, mockScheduleClient := setup(t)
		defer ctrl.Finish()

		lessonID := uuid.New()
		fileID := uuid.New()
		receiptID := uuid.New()

		mockScheduleClient.EXPECT().GetLesson(gomock.Any(), &api.GetLessonRequest{
			Id: lessonID.String(),
		}).Return(&api.Lesson{
			Id:             lessonID.String(),
			ConnectionLink: proto.String("nigger"),    // вместо &variable
			PriceRub:       proto.Int32(100),          // вместо &priceRub
			PaymentInfo:    proto.String("some info"), // вместо &paymentInfo
		}, nil)

		mockScheduleClient.EXPECT().UpdateLesson(gomock.Any(), &api.UpdateLessonRequest{
			Id:             lessonID.String(),
			ConnectionLink: proto.String("nigger"),
			PriceRub:       proto.Int32(100),
			PaymentInfo:    proto.String("some info"),
		}).Return(&api.Lesson{}, nil)

		mockRepo.EXPECT().ExistsByID(gomock.Any(), gomock.Any()).Return(false, nil)

		mockRepo.EXPECT().CreateReceipt(gomock.Any(), &models.PaymentReceiptCreateInput{
			LessonID:   lessonID,
			FileID:     fileID,
			IsVerified: true,
		}).Return(&models.PaymentReceipt{
			ID:         receiptID,
			LessonID:   lessonID,
			FileID:     fileID,
			IsVerified: true,
		}, nil)

		result, err := svc.SubmitPaymentReceipt(context.Background(), &models.SubmitPaymentReceiptInput{
			LessonId: lessonID,
			FileId:   fileID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != receiptID {
			t.Fatalf("expected receipt ID %v, got %v", receiptID, result.ID)
		}
	})

	t.Run("Error_InvalidInput", func(t *testing.T) {
		_, svc, _, _, _, _ := setup(t)

		testCases := []struct {
			name  string
			input *models.SubmitPaymentReceiptInput
		}{
			{"EmptyLessonID", &models.SubmitPaymentReceiptInput{FileId: uuid.New()}},
			{"EmptyFileID", &models.SubmitPaymentReceiptInput{LessonId: uuid.New()}},
			{"BothEmpty", &models.SubmitPaymentReceiptInput{}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := svc.SubmitPaymentReceipt(context.Background(), tc.input)
				if !errors.Is(err, errdefs.ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument, got %v", err)
				}
			})
		}
	})

	t.Run("Error_LessonAlreadyPaid", func(t *testing.T) {
		ctrl, svc, _, _, _, mockScheduleClient := setup(t)
		defer ctrl.Finish()

		lessonID := uuid.New()
		mockScheduleClient.EXPECT().GetLesson(gomock.Any(), gomock.Any()).Return(&api.Lesson{
			IsPaid: true,
		}, nil)

		_, err := svc.SubmitPaymentReceipt(context.Background(), &models.SubmitPaymentReceiptInput{
			LessonId: lessonID,
			FileId:   uuid.New(),
		})
		if !errors.Is(err, errdefs.ErrAlreadyExists) {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("Error_UpdateLesson", func(t *testing.T) {
		ctrl, svc, _, _, _, mockSchedule := setup(t)
		defer ctrl.Finish()

		mockSchedule.EXPECT().
			GetLesson(gomock.Any(), gomock.Any()).
			Return(&api.Lesson{IsPaid: false}, nil)
		mockSchedule.EXPECT().
			UpdateLesson(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("upd error"))

		_, err := svc.SubmitPaymentReceipt(context.Background(), &models.SubmitPaymentReceiptInput{
			LessonId: uuid.New(), FileId: uuid.New(),
		})
		if err == nil || err.Error() != "upd error" {
			t.Fatalf("want upd error, got %v", err)
		}
	})
	t.Run("Error_CreateReceipt", func(t *testing.T) {
		ctrl, svc, mockRepo, _, _, mockSchedule := setup(t)
		defer ctrl.Finish()

		mockSchedule.EXPECT().GetLesson(gomock.Any(), gomock.Any()).Return(&api.Lesson{PriceRub: proto.Int32(1)}, nil)
		mockSchedule.EXPECT().UpdateLesson(gomock.Any(), gomock.Any()).Return(&api.Lesson{}, nil)
		mockRepo.EXPECT().ExistsByID(gomock.Any(), gomock.Any()).Return(false, nil)
		mockRepo.EXPECT().
			CreateReceipt(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("db error"))

		_, err := svc.SubmitPaymentReceipt(context.Background(), &models.SubmitPaymentReceiptInput{
			LessonId: uuid.New(), FileId: uuid.New(),
		})
		if err == nil || err.Error() != errdefs.ErrNotFound.Error() {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
	t.Run("RetryLogic_SucceedsAfterRetries", func(t *testing.T) {
		ctrl, svc, mockRepo, _, _, mockScheduleClient := setup(t)
		defer ctrl.Finish()

		lessonID := uuid.New()
		fileID := uuid.New()

		mockScheduleClient.EXPECT().GetLesson(gomock.Any(), gomock.Any()).
			Return(&api.Lesson{Id: lessonID.String(), IsPaid: false}, nil)
		mockScheduleClient.EXPECT().UpdateLesson(gomock.Any(), gomock.Any()).
			Return(&api.Lesson{}, nil)

		mockRepo.EXPECT().ExistsByID(gomock.Any(), gomock.Any()).
			Return(false, nil)

		retriableError := status.Error(codes.Unavailable, "service down")
		mockRepo.EXPECT().CreateReceipt(gomock.Any(), gomock.Any()).
			Return(nil, retriableError).Times(4)
		mockRepo.EXPECT().CreateReceipt(gomock.Any(), gomock.Any()).
			Return(&models.PaymentReceipt{}, nil).Times(1)

		_, err := svc.SubmitPaymentReceipt(context.Background(), &models.SubmitPaymentReceiptInput{
			LessonId: lessonID,
			FileId:   fileID,
		})

		assert.NoError(t, err)
	})

}

func TestGetPaymentInfo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl, svc, _, _, _, mockScheduleClient := setup(t)
		defer ctrl.Finish()

		lessonID := uuid.New()
		input := &models.GetPaymentInfoInput{LessonId: lessonID}

		mockScheduleClient.EXPECT().GetLesson(gomock.Any(), &api.GetLessonRequest{
			Id: lessonID.String(),
		}).Return(&api.Lesson{
			Id:          lessonID.String(),
			PriceRub:    proto.Int32(1500),
			PaymentInfo: proto.String("Payment instructions"),
		}, nil)

		info, err := svc.GetPaymentInfo(context.Background(), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info == nil || info.LessonID != lessonID || info.PriceRUB != 1500 {
			t.Fatal("invalid payment info returned")
		}
	})

	t.Run("Error_InvalidInput", func(t *testing.T) {
		_, svc, _, _, _, _ := setup(t)

		_, err := svc.GetPaymentInfo(context.Background(), &models.GetPaymentInfoInput{})
		if err == nil {
			t.Fatal("expected error for empty lesson ID")
		}
	})

	t.Run("Error_LessonNotFound", func(t *testing.T) {
		ctrl, svc, _, _, _, mockScheduleClient := setup(t)
		defer ctrl.Finish()

		lessonID := uuid.New()
		mockScheduleClient.EXPECT().GetLesson(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))

		_, err := svc.GetPaymentInfo(context.Background(), &models.GetPaymentInfoInput{LessonId: lessonID})
		if err == nil {
			t.Fatal("expected error when lesson not found")
		}
	})
	t.Run("Error_ReceiptNotFound", func(t *testing.T) {
		ctrl, svc, mockRepo, _, _, _ := setup(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().
			GetReceiptByID(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("not found"))

		_, err := svc.GetReceiptFile(context.Background(), &models.GetReceiptFileInput{ReceiptId: uuid.New()})
		if err == nil {
			t.Fatal("expected error when receipt missing")
		}
	})

}

func TestGetReceipt(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl, svc, mockRepo, _, _, _ := setup(t)
		defer ctrl.Finish()

		receiptID := uuid.New()
		input := &models.GetReceiptInput{ReceiptId: receiptID}

		receipt := &models.PaymentReceipt{
			ID:         receiptID,
			LessonID:   uuid.New(),
			FileID:     uuid.New(),
			IsVerified: true,
			CreatedAt:  time.Now(),
			EditedAt:   time.Now(),
		}

		mockRepo.EXPECT().GetReceiptByID(gomock.Any(), receiptID).Return(receipt, nil)

		result, err := svc.GetReceipt(context.Background(), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil || result.ID != receiptID {
			t.Fatal("invalid receipt returned")
		}
	})

	t.Run("Error_InvalidInput", func(t *testing.T) {
		_, svc, _, _, _, _ := setup(t)

		_, err := svc.GetReceipt(context.Background(), &models.GetReceiptInput{})
		if err == nil {
			t.Fatal("expected error for empty receipt ID")
		}
	})

	t.Run("Error_ReceiptNotFound", func(t *testing.T) {
		ctrl, svc, mockRepo, _, _, _ := setup(t)
		defer ctrl.Finish()

		receiptID := uuid.New()
		mockRepo.EXPECT().GetReceiptByID(gomock.Any(), receiptID).Return(nil, errors.New("not found"))

		_, err := svc.GetReceipt(context.Background(), &models.GetReceiptInput{ReceiptId: receiptID})
		if err == nil {
			t.Fatal("expected error when receipt not found")
		}
	})
}

func TestVerifyReceipt(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl, svc, mockRepo, _, _, _ := setup(t)
		defer ctrl.Finish()

		receiptID := uuid.New()
		input := &models.VerifyReceiptInput{ReceiptId: receiptID}

		updatedReceipt := &models.PaymentReceipt{
			ID:         receiptID,
			IsVerified: true,
		}

		mockRepo.EXPECT().UpdateReceipt(gomock.Any(), receiptID, true).Return(updatedReceipt, nil)

		result, err := svc.VerifyReceipt(context.Background(), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil || !result.IsVerified {
			t.Fatal("receipt not verified")
		}
	})

	t.Run("Error_InvalidInput", func(t *testing.T) {
		_, svc, _, _, _, _ := setup(t)

		_, err := svc.VerifyReceipt(context.Background(), &models.VerifyReceiptInput{})
		if err == nil {
			t.Fatal("expected error for empty receipt ID")
		}
	})

	t.Run("Error_ReceiptNotFound", func(t *testing.T) {
		ctrl, svc, mockRepo, _, _, _ := setup(t)
		defer ctrl.Finish()

		receiptID := uuid.New()
		mockRepo.EXPECT().UpdateReceipt(gomock.Any(), receiptID, true).Return(nil, errors.New("not found"))

		_, err := svc.VerifyReceipt(context.Background(), &models.VerifyReceiptInput{ReceiptId: receiptID})
		if err == nil {
			t.Fatal("expected error when receipt not found")
		}
	})
}

func TestGetReceiptFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl, svc, mockRepo, _, mockFileClient, _ := setup(t)
		defer ctrl.Finish()

		receiptID := uuid.New()
		fileID := uuid.New()
		input := &models.GetReceiptFileInput{ReceiptId: receiptID}

		receipt := &models.PaymentReceipt{
			ID:         receiptID,
			FileID:     fileID,
			IsVerified: true,
		}

		mockRepo.EXPECT().GetReceiptByID(gomock.Any(), receiptID).Return(receipt, nil)
		mockFileClient.EXPECT().GenerateDownloadURL(gomock.Any(), &api2.GenerateDownloadURLRequest{
			FileId: fileID.String(),
		}).Return(&api2.DownloadURL{
			Url: "https://storage.example.com/file123",
		}, nil)

		result, err := svc.GetReceiptFile(context.Background(), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil || result.URL != "https://storage.example.com/file123" {
			t.Fatal("invalid download URL returned")
		}
	})

	t.Run("Error_InvalidInput", func(t *testing.T) {
		_, svc, _, _, _, _ := setup(t)

		_, err := svc.GetReceiptFile(context.Background(), &models.GetReceiptFileInput{})
		if err == nil {
			t.Fatal("expected error for empty receipt ID")
		}
	})

	t.Run("Error_FileServiceUnavailable", func(t *testing.T) {
		ctrl, svc, mockRepo, _, mockFileClient, _ := setup(t)
		defer ctrl.Finish()

		receiptID := uuid.New()
		fileID := uuid.New()
		receipt := &models.PaymentReceipt{
			ID:         receiptID,
			FileID:     fileID,
			IsVerified: true,
		}

		mockRepo.EXPECT().GetReceiptByID(gomock.Any(), receiptID).Return(receipt, nil)
		mockFileClient.EXPECT().GenerateDownloadURL(gomock.Any(), gomock.Any()).Return(nil, errors.New("service unavailable"))

		_, err := svc.GetReceiptFile(context.Background(), &models.GetReceiptFileInput{ReceiptId: receiptID})
		if err == nil {
			t.Fatal("expected error when file service unavailable")
		}
	})
}
