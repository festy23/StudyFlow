//go:generate mockgen -source=handler.go -destination=../mocks/payment_service_mocks.go -package=mocks

package handler

import (
	"common_library/logging"
	"context"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	errdefs "paymentservice/internal/errors"
	"paymentservice/internal/models"
	pb "paymentservice/pkg/api"
	"slices"
)

type PaymentService interface {
	GetPaymentInfo(ctx context.Context, input *models.GetPaymentInfoInput) (*models.PaymentInfo, error)
	SubmitPaymentReceipt(ctx context.Context, input *models.SubmitPaymentReceiptInput) (*models.PaymentReceipt, error)
	GetReceipt(ctx context.Context, input *models.GetReceiptInput) (*models.PaymentReceipt, error)
	VerifyReceipt(ctx context.Context, input *models.VerifyReceiptInput) (*models.PaymentReceipt, error)
	GetReceiptFile(ctx context.Context, input *models.GetReceiptFileInput) (*models.ReceiptFileUrl, error)
}

type PaymentServiceServer struct {
	pb.UnimplementedPaymentServiceServer
	service PaymentService
}

func NewPaymentServiceServer(paymentService PaymentService) *PaymentServiceServer {
	return &PaymentServiceServer{service: paymentService}
}
func (h *PaymentServiceServer) GetPaymentInfo(ctx context.Context, req *pb.GetPaymentInfoRequest) (*pb.PaymentInfo, error) {
	lessonID, err := uuid.Parse(*req.LessonId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid LessonId: %v", err)
	}

	input := &models.GetPaymentInfoInput{LessonId: lessonID}
	paymentInfo, err := h.service.GetPaymentInfo(ctx, input)
	if err != nil {
		if errors.Is(err, errdefs.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "payment info not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "internal server error: %v", err)
	}

	return &pb.PaymentInfo{
		LessonId:    req.LessonId,
		PriceRub:    &paymentInfo.PriceRUB,
		PaymentInfo: &paymentInfo.PaymentDetails,
	}, nil
}

func (h *PaymentServiceServer) SubmitPaymentReceipt(ctx context.Context, req *pb.SubmitPaymentReceiptRequest) (*pb.Receipt, error) {
	lessonID, err := uuid.Parse(*req.LessonId)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid lesson ID: "+err.Error()).Err()
	}
	fileID, err := uuid.Parse(*req.FileId)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid file ID: "+err.Error()).Err()
	}

	input := &models.SubmitPaymentReceiptInput{
		LessonId: lessonID,
		FileId:   fileID,
	}
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Info(ctx, "submitting payment receipt", zap.Any("input", input))
	}
	paymentReceipt, err := h.service.SubmitPaymentReceipt(ctx, input)
	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ErrPermissionDenied, errdefs.ErrInvalidArgument)
	}
	return toPbReceipt(paymentReceipt), nil
}

func (h *PaymentServiceServer) GetReceipt(ctx context.Context, req *pb.GetReceiptRequest) (*pb.Receipt, error) {
	receiptID, err := uuid.Parse(req.ReceiptId)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid receipt ID: "+err.Error()).Err()
	}

	input := &models.GetReceiptInput{
		ReceiptId: receiptID,
	}
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Info(ctx, "getting payment receipt", zap.Any("input", input))
	}
	paymentReceipt, err := h.service.GetReceipt(ctx, input)
	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ErrPermissionDenied)
	}
	return toPbReceipt(paymentReceipt), nil
}

func (h *PaymentServiceServer) VerifyReceipt(ctx context.Context, req *pb.VerifyReceiptRequest) (*pb.Receipt, error) {
	receiptID, err := uuid.Parse(req.ReceiptId)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid receipt ID: "+err.Error()).Err()
	}

	input := &models.VerifyReceiptInput{
		ReceiptId: receiptID,
	}
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Info(ctx, "verifying receipt", zap.Any("input", input))
	}
	paymentReceipt, err := h.service.VerifyReceipt(ctx, input)
	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ErrPermissionDenied)
	}
	return toPbReceipt(paymentReceipt), nil
}

func (h *PaymentServiceServer) GetReceiptFile(ctx context.Context, req *pb.GetReceiptFileRequest) (*pb.ReceiptFileURL, error) {
	receiptID, err := uuid.Parse(req.ReceiptId)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid receipt ID: "+err.Error()).Err()
	}

	input := &models.GetReceiptFileInput{
		ReceiptId: receiptID,
	}
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Info(ctx, "getting receipt file", zap.Any("input", input))
	}
	receiptFileURL, err := h.service.GetReceiptFile(ctx, input)
	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ErrInvalidArgument, errdefs.ErrPermissionDenied)
	}

	return toPbReceiptFileURl(receiptFileURL), nil
}

func toPbReceiptFileURl(receiptFileUrl *models.ReceiptFileUrl) *pb.ReceiptFileURL {
	return &pb.ReceiptFileURL{
		Url: &receiptFileUrl.URL,
	}
}
func toPbReceipt(receipt *models.PaymentReceipt) *pb.Receipt {
	lessonID := receipt.LessonID.String()
	fileID := receipt.FileID.String()

	return &pb.Receipt{
		Id:         receipt.ID.String(),
		LessonId:   &lessonID,
		FileId:     &fileID,
		IsVerified: receipt.IsVerified,
		CreatedAt:  timestamppb.New(receipt.CreatedAt),
		EditedAt:   timestamppb.New(receipt.EditedAt),
	}
}

//func toPbPaymentInfo(paymentInfo *models.PaymentInfo) *pb.PaymentInfo {
//	return &pb.PaymentInfo{
//		LessonId:    paymentInfo.LessonID.String(),
//		PriceRub:    paymentInfo.PriceRUB,
//		PaymentInfo: paymentInfo.PaymentDetails,
//	}
//}

func mapError(err error, possibleErrors ...error) error {
	switch {

	case err == nil:
		return nil

	case errors.Is(err, errdefs.ErrNotFound) && slices.Contains(possibleErrors, errdefs.ErrNotFound):
		return status.New(codes.NotFound, "resource not found").Err()

	case errors.Is(err, errdefs.ErrPermissionDenied) && slices.Contains(possibleErrors, errdefs.ErrPermissionDenied):
		return status.New(codes.PermissionDenied, "permission denied").Err()

	case errors.Is(err, errdefs.ErrInvalidPayment) && slices.Contains(possibleErrors, errdefs.ErrInvalidPayment):
		return status.New(codes.Unauthenticated, "invalid payment").Err()

	case errors.Is(err, errdefs.ErrInvalidArgument) && slices.Contains(possibleErrors, errdefs.ErrInvalidArgument):
		return status.New(codes.InvalidArgument, "invalid argument provided").Err()

	default:
		return status.New(codes.Internal, "internal server error").Err()
	}
}
