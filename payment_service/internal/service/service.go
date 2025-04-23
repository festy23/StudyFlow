//go:generate mockgen -source=service.go -destination=../mocks/payment_mocks.go -package=mocks

package service

import (
	"context"
	api2 "fileservice/pkg/api"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"math"
	"paymentservice/internal/clients"
	errdefs "paymentservice/internal/errors"
	"paymentservice/internal/models"
	api3 "schedule_service/pkg/api"
	"time"
)

const maxRetries = 6                      // Максимальное количество попыток
const retryDelay = 100 * time.Millisecond // Задержка между попытками

type IPaymentRepo interface {
	CreateReceipt(ctx context.Context, receipt *models.PaymentReceiptCreateInput) (*models.PaymentReceipt, error)

	GetReceiptByID(ctx context.Context, id uuid.UUID) (*models.PaymentReceipt, error)

	UpdateReceipt(ctx context.Context, id uuid.UUID, isVerified bool) (*models.PaymentReceipt, error)

	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)

	GetReceiptByLessonID(ctx context.Context, lessonID uuid.UUID) (*models.PaymentReceipt, error)
}

type PaymentService struct {
	repo           IPaymentRepo
	userClient     clients.UserServiceClient
	fileClient     clients.FileServiceClient
	scheduleClient clients.ScheduleServiceClient
}

func NewPaymentService(
	repo IPaymentRepo,
	userClient clients.UserServiceClient,
	fileClient clients.FileServiceClient,
	scheduleClient clients.ScheduleServiceClient,
) *PaymentService {

	return &PaymentService{
		repo:           repo,
		userClient:     userClient,
		fileClient:     fileClient,
		scheduleClient: scheduleClient,
	}
}

func (s *PaymentService) SubmitPaymentReceipt(ctx context.Context, input *models.SubmitPaymentReceiptInput) (*models.PaymentReceipt, error) {
	if input.FileId == uuid.Nil || input.LessonId == uuid.Nil {
		return nil, errdefs.ErrInvalidArgument
	}

	getLessonRequest := &api3.GetLessonRequest{
		Id: input.LessonId.String(),
	}

	lesson, err := retry[*api3.Lesson](ctx, maxRetries, retryDelay, func() (*api3.Lesson, error) {
		return s.scheduleClient.GetLesson(ctx, getLessonRequest)
	})
	if err != nil {
		return nil, err
	}

	if lesson.IsPaid {
		return nil, errdefs.ErrAlreadyExists
	}

	lesson.IsPaid = true
	updateLessonRequest := &api3.UpdateLessonRequest{
		Id:             lesson.Id,
		ConnectionLink: lesson.ConnectionLink,
		PriceRub:       lesson.PriceRub,
		PaymentInfo:    lesson.PaymentInfo,
	}
	lesson, err = retry[*api3.Lesson](ctx, maxRetries, retryDelay, func() (*api3.Lesson, error) {
		return s.scheduleClient.UpdateLesson(ctx, updateLessonRequest)
	})
	if err != nil {
		return nil, err
	}
	newReceiptID := uuid.New()
	exists, err := s.repo.ExistsByID(ctx, newReceiptID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errdefs.ErrAlreadyExists
	}

	createReceiptInput := &models.PaymentReceiptCreateInput{
		LessonID:   input.LessonId,
		FileID:     input.FileId,
		IsVerified: true,
	}
	receipt, err := retry(ctx, maxRetries, retryDelay, func() (*models.PaymentReceipt, error) {
		return s.repo.CreateReceipt(ctx, createReceiptInput)
	})
	if err != nil {
		return nil, errdefs.ErrNotFound
	}

	// отправить ивент уведомление

	return receipt, nil
}

func (s *PaymentService) GetPaymentInfo(ctx context.Context, input *models.GetPaymentInfoInput) (*models.PaymentInfo, error) {
	if input.LessonId == uuid.Nil {
		return nil, errdefs.ErrInvalidArgument
	}

	getLessonRequest := &api3.GetLessonRequest{
		Id: input.LessonId.String(),
	}

	lesson, err := retry(ctx, maxRetries, retryDelay, func() (*api3.Lesson, error) {
		return s.scheduleClient.GetLesson(ctx, getLessonRequest)
	})
	if err != nil {
		return nil, err
	}
	if *lesson.PriceRub == 0 {
		*lesson.PriceRub = 0
	}
	paymentInfo := &models.PaymentInfo{
		LessonID:       input.LessonId,
		PriceRUB:       *lesson.PriceRub,
		PaymentDetails: *lesson.PaymentInfo,
	}
	return paymentInfo, nil
}
func (s *PaymentService) GetReceipt(ctx context.Context, input *models.GetReceiptInput) (*models.PaymentReceipt, error) {
	if input.ReceiptId == uuid.Nil {
		return nil, errdefs.ErrInvalidArgument
	}

	receipt, err := retry[*models.PaymentReceipt](ctx, maxRetries, retryDelay, func() (*models.PaymentReceipt, error) {
		return s.repo.GetReceiptByID(ctx, input.ReceiptId)
	})
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (s *PaymentService) VerifyReceipt(ctx context.Context, input *models.VerifyReceiptInput) (*models.PaymentReceipt, error) {
	if input.ReceiptId == uuid.Nil {
		return nil, errdefs.ErrInvalidArgument
	}
	receipt, err := retry(ctx, maxRetries, retryDelay, func() (*models.PaymentReceipt, error) {
		return s.repo.UpdateReceipt(ctx, input.ReceiptId, true)
	})
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (s *PaymentService) GetReceiptFile(ctx context.Context, input *models.GetReceiptFileInput) (*models.ReceiptFileUrl, error) {
	if input.ReceiptId == uuid.Nil {
		return nil, errdefs.ErrInvalidArgument
	}
	receipt, err := retry[*models.PaymentReceipt](ctx, maxRetries, retryDelay, func() (*models.PaymentReceipt, error) {
		return s.repo.GetReceiptByID(ctx, input.ReceiptId)
	})
	if err != nil {
		return nil, err
	}
	generateDownloadURLRequest := &api2.GenerateDownloadURLRequest{FileId: receipt.FileID.String()}
	url, err := s.fileClient.GenerateDownloadURL(ctx, generateDownloadURLRequest)
	if err != nil {
		return nil, err
	}
	receiptFileURL := &models.ReceiptFileUrl{
		URL: url.GetUrl(),
	}
	return receiptFileURL, nil
}

func retry[T any](
	ctx context.Context,
	attempts int,
	baseDelay time.Duration,
	fn func() (T, error),
) (T, error) {
	var zero T
	var err error

	for i := 0; i < attempts; i++ {
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
			result, fnErr := fn()
			if fnErr == nil {
				return result, nil
			}
			err = fnErr

			if !isRetriable(fnErr) {
				return zero, fnErr
			}

			delay := time.Duration(math.Pow(2, float64(i))) * baseDelay
			log.Printf("Retrying after error: %v (attempt %d)", fnErr, i+1)
			time.Sleep(delay)
		}
	}
	return zero, fmt.Errorf("after %d attempts: %w", attempts, err)
}

func isRetriable(err error) bool {
	if s, ok := status.FromError(err); ok {
		return s.Code() == codes.NotFound || s.Code() == codes.PermissionDenied ||
			s.Code() == codes.Unavailable || s.Code() == codes.InvalidArgument || s.Code() == codes.Internal
	}
	return false
}
