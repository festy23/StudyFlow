package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"payment_service/internal/service"
	paymentv1 "payment_service/proto"
)

type handler struct {
	paymentv1.UnimplementedPaymentServiceServer
	service *service.PaymentService
}

func newHandler(svc *service.PaymentService) *handler {
	return &handler{service: svc}
}

func (h *handler) GetPaymentInfo(ctx context.Context, req *paymentv1.GetPaymentInfoRequest) (*paymentv1.PaymentInfo, error) {
	info, err := h.service.GetPaymentInfo(ctx, req.LessonId)
	if err != nil {
		switch err {
		case service.ErrNotFound:
			return nil, status.Error(codes.NotFound, "lesson not found")
		case service.ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &paymentv1.PaymentInfo{
		LessonId:    info.LessonID,
		PriceRub:    info.PriceRub,
		PaymentInfo: info.PaymentInfo,
	}, nil
}

func (h *handler) SubmitPaymentReceipt(ctx context.Context, req *paymentv1.SubmitPaymentReceiptRequest) (*paymentv1.Receipt, error) {
	userID, ok := getUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user_id not found")
	}

	// Проверка что пользователь - ученик этого урока
	// Реализация проверки через schedule_service

	receipt, err := h.service.SubmitPaymentReceipt(ctx, req.LessonId, req.FileId, userID)
	if err != nil {
		switch err {
		case service.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, "invalid arguments")
		case service.ErrNotFound:
			return nil, status.Error(codes.NotFound, "lesson not found")
		case service.ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &paymentv1.Receipt{
		Id:         receipt.ID,
		LessonId:   receipt.LessonID,
		FileId:     receipt.FileID,
		IsVerified: receipt.IsVerified,
		CreatedAt:  timestamppb.New(receipt.CreatedAt),
		EditedAt:   timestamppb.New(receipt.EditedAt),
	}, nil
}

func (h *handler) GetReceipt(ctx context.Context, req *paymentv1.GetReceiptRequest) (*paymentv1.Receipt, error) {
	userID, ok := getUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user_id not found")
	}

	receipt, err := h.service.GetReceipt(ctx, req.ReceiptId, userID)
	if err != nil {
		switch err {
		case service.ErrNotFound:
			return nil, status.Error(codes.NotFound, "receipt not found")
		case service.ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &paymentv1.Receipt{
		Id:         receipt.ID,
		LessonId:   receipt.LessonID,
		FileId:     receipt.FileID,
		IsVerified: receipt.IsVerified,
		CreatedAt:  timestamppb.New(receipt.CreatedAt),
		EditedAt:   timestamppb.New(receipt.EditedAt),
	}, nil
}

func (h *handler) VerifyReceipt(ctx context.Context, req *paymentv1.VerifyReceiptRequest) (*paymentv1.Receipt, error) {
	userID, ok := getUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user_id not found")
	}

	// Проверка что пользователь - репетитор этого урока
	// Реализация проверки через schedule_service

	receipt, err := h.service.VerifyReceipt(ctx, req.ReceiptId, userID)
	if err != nil {
		switch err {
		case service.ErrNotFound:
			return nil, status.Error(codes.NotFound, "receipt not found")
		case service.ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &paymentv1.Receipt{
		Id:         receipt.ID,
		LessonId:   receipt.LessonID,
		FileId:     receipt.FileID,
		IsVerified: receipt.IsVerified,
		CreatedAt:  timestamppb.New(receipt.CreatedAt),
		EditedAt:   timestamppb.New(receipt.EditedAt),
	}, nil
}

func (h *handler) GetReceiptFile(ctx context.Context, req *paymentv1.GetReceiptFileRequest) (*paymentv1.ReceiptFileURL, error) {
	userID, ok := getUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user_id not found")
	}

	url, err := h.service.GetReceiptFile(ctx, req.ReceiptId, userID)
	if err != nil {
		switch err {
		case service.ErrInvalidArgument:
			return nil, status.Error(codes.InvalidArgument, "invalid receipt_id")
		case service.ErrNotFound:
			return nil, status.Error(codes.NotFound, "receipt or lesson not found")
		case service.ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &paymentv1.ReceiptFileURL{
		Url: url,
	}, nil
}
