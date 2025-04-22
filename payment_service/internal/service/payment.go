package service

import (
	"context"
	"errors"

	"payment_service/internal/domain"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidArgument  = errors.New("invalid argument")
)

type PaymentService struct {
	repo        domain.PaymentRepository
	scheduleSvc ScheduleServiceClient
	userSvc     UserServiceClient
	fileSvc     FileServiceClient
	notifier    NotificationService
}

func NewPaymentService(
	repo domain.PaymentRepository,
	scheduleSvc ScheduleServiceClient,
	userSvc UserServiceClient,
	fileSvc FileServiceClient,
	notifier NotificationService,
) *PaymentService {
	return &PaymentService{
		repo:        repo,
		scheduleSvc: scheduleSvc,
		userSvc:     userSvc,
		fileSvc:     fileSvc,
		notifier:    notifier,
	}
}

func (s *PaymentService) GetPaymentInfo(ctx context.Context, lessonID string) (*domain.PaymentInfo, error) {
	// 1. Проверить права доступа (должен быть ученик из урока)
	// 2. Получить информацию об уроке из schedule_service
	// 3. Получить информацию о цене из user_service (если нет в уроке)
	// 4. Вернуть PaymentInfo
	return s.repo.GetPaymentInfo(ctx, lessonID)
}

func (s *PaymentService) SubmitPaymentReceipt(ctx context.Context, lessonID, fileID string) (*domain.Receipt, error) {
	// 1. Проверить права доступа (должен быть ученик из урока)
	// 2. Создать запись чека
	// 3. Отправить уведомление об оплате
	receipt := &domain.Receipt{
		LessonID:   lessonID,
		FileID:     fileID,
		IsVerified: false,
	}
	err := s.repo.CreateReceipt(ctx, receipt)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (s *PaymentService) GetReceipt(ctx context.Context, receiptID string) (*domain.Receipt, error) {
	// 1. Получить чек
	// 2. Проверить права доступа (должен быть участник урока)
	return s.repo.GetReceipt(ctx, receiptID)
}

func (s *PaymentService) VerifyReceipt(ctx context.Context, receiptID string) (*domain.Receipt, error) {
	// 1. Проверить права доступа (должен быть репетитор из урока)
	// 2. Обновить статус верификации
	err := s.repo.UpdateReceiptVerification(ctx, receiptID, true)
	if err != nil {
		return nil, err
	}
	return s.repo.GetReceipt(ctx, receiptID)
}
