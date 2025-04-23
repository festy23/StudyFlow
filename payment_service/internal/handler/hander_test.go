package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	errdefs "paymentservice/internal/errors"
	"paymentservice/internal/mocks"
	"paymentservice/internal/models"
	pb "paymentservice/pkg/api"
	"testing"
	"time"
)

func TestGetPaymentInfo_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lessonID := uuid.New()
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.GetPaymentInfoInput{LessonId: lessonID}
	response := &models.PaymentInfo{
		LessonID:       lessonID,
		PriceRUB:       1000,
		PaymentDetails: "test payment details",
	}
	mockSvc.EXPECT().GetPaymentInfo(ctx, input).Return(response, nil)
	lid := lessonID.String()
	res, err := h.GetPaymentInfo(ctx, &pb.GetPaymentInfoRequest{LessonId: &lid})
	assert.NoError(t, err)
	assert.Equal(t, lessonID.String(), *res.LessonId)         // Разыменовываем указатель
	assert.Equal(t, int32(1000), *res.PriceRub)               // Разыменовываем указатель
	assert.Equal(t, "test payment details", *res.PaymentInfo) // Разыменовываем указатель
}

func TestGetPaymentInfo_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lessonID := uuid.New()
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.GetPaymentInfoInput{LessonId: lessonID}
	mockSvc.EXPECT().GetPaymentInfo(ctx, input).Return(nil, errdefs.ErrNotFound)
	lid := lessonID.String()
	_, err := h.GetPaymentInfo(ctx, &pb.GetPaymentInfoRequest{LessonId: &lid})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestGetPaymentInfo_InvalidLessonID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	str := "invalid-uuid"
	_, err := h.GetPaymentInfo(ctx, &pb.GetPaymentInfoRequest{LessonId: &str})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSubmitPaymentReceipt_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lessonID := uuid.New()
	fileID := uuid.New()
	createdAt := time.Now().UTC().Truncate(time.Second)
	editedAt := createdAt
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.SubmitPaymentReceiptInput{LessonId: lessonID, FileId: fileID}
	response := &models.PaymentReceipt{
		ID:         uuid.New(),
		LessonID:   lessonID,
		FileID:     fileID,
		IsVerified: false,
		CreatedAt:  createdAt,
		EditedAt:   editedAt,
	}
	mockSvc.EXPECT().SubmitPaymentReceipt(ctx, input).Return(response, nil)
	lID := lessonID.String()
	fID := fileID.String()
	res, err := h.SubmitPaymentReceipt(ctx, &pb.SubmitPaymentReceiptRequest{LessonId: &lID, FileId: &fID})
	assert.NoError(t, err)
	assert.Equal(t, response.ID.String(), res.Id)
	assert.Equal(t, lessonID.String(), *res.LessonId)
	assert.Equal(t, fileID.String(), *res.FileId)
	assert.Equal(t, false, res.IsVerified)
	assert.Equal(t, createdAt, res.CreatedAt.AsTime().Truncate(time.Second))
	assert.Equal(t, editedAt, res.EditedAt.AsTime().Truncate(time.Second))
}

func TestSubmitPaymentReceipt_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lessonID := uuid.New()
	fileID := uuid.New()
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.SubmitPaymentReceiptInput{LessonId: lessonID, FileId: fileID}
	mockSvc.EXPECT().SubmitPaymentReceipt(ctx, input).Return(nil, errdefs.ErrNotFound)
	lID := lessonID.String()
	fID := fileID.String()
	_, err := h.SubmitPaymentReceipt(ctx, &pb.SubmitPaymentReceiptRequest{LessonId: &lID, FileId: &fID})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestSubmitPaymentReceipt_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lessonID := uuid.New()
	fileID := uuid.New()
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.SubmitPaymentReceiptInput{LessonId: lessonID, FileId: fileID}
	mockSvc.EXPECT().SubmitPaymentReceipt(ctx, input).Return(nil, errdefs.ErrPermissionDenied)
	lID := lessonID.String()
	fID := fileID.String()
	_, err := h.SubmitPaymentReceipt(ctx, &pb.SubmitPaymentReceiptRequest{LessonId: &lID, FileId: &fID})
	assert.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestSubmitPaymentReceipt_InvalidLessonID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	newID := uuid.New().String()
	str := "invalid-uuid"
	_, err := h.SubmitPaymentReceipt(ctx, &pb.SubmitPaymentReceiptRequest{LessonId: &str, FileId: &newID})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSubmitPaymentReceipt_InvalidFileID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	newID := uuid.New().String()
	str := "invalid-uuid"
	_, err := h.SubmitPaymentReceipt(ctx, &pb.SubmitPaymentReceiptRequest{LessonId: &newID, FileId: &str})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetReceipt_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiptID := uuid.New()
	lessonID := uuid.New()
	fileID := uuid.New()
	createdAt := time.Now().UTC().Truncate(time.Second)
	editedAt := createdAt
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.GetReceiptInput{ReceiptId: receiptID}
	response := &models.PaymentReceipt{
		ID:         receiptID,
		LessonID:   lessonID,
		FileID:     fileID,
		IsVerified: true,
		CreatedAt:  createdAt,
		EditedAt:   editedAt,
	}
	mockSvc.EXPECT().GetReceipt(ctx, input).Return(response, nil)
	res, err := h.GetReceipt(ctx, &pb.GetReceiptRequest{ReceiptId: receiptID.String()})
	assert.NoError(t, err)
	assert.Equal(t, receiptID.String(), res.Id)
	assert.Equal(t, lessonID.String(), *res.LessonId)
	assert.Equal(t, fileID.String(), *res.FileId)
	assert.Equal(t, true, res.IsVerified)
	assert.Equal(t, createdAt, res.CreatedAt.AsTime().Truncate(time.Second))
	assert.Equal(t, editedAt, res.EditedAt.AsTime().Truncate(time.Second))
}

func TestGetReceipt_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiptID := uuid.New()
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.GetReceiptInput{ReceiptId: receiptID}
	mockSvc.EXPECT().GetReceipt(ctx, input).Return(nil, errdefs.ErrNotFound)
	_, err := h.GetReceipt(ctx, &pb.GetReceiptRequest{ReceiptId: receiptID.String()})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestGetReceipt_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiptID := uuid.New()
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.GetReceiptInput{ReceiptId: receiptID}
	mockSvc.EXPECT().GetReceipt(ctx, input).Return(nil, errdefs.ErrPermissionDenied)
	_, err := h.GetReceipt(ctx, &pb.GetReceiptRequest{ReceiptId: receiptID.String()})
	assert.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestGetReceipt_InvalidReceiptID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	_, err := h.GetReceipt(ctx, &pb.GetReceiptRequest{ReceiptId: "invalid-uuid"})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestVerifyReceipt_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiptID := uuid.New()
	lessonID := uuid.New()
	fileID := uuid.New()
	createdAt := time.Now().UTC().Truncate(time.Second)
	editedAt := createdAt
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.VerifyReceiptInput{ReceiptId: receiptID}
	response := &models.PaymentReceipt{
		ID:         receiptID,
		LessonID:   lessonID,
		FileID:     fileID,
		IsVerified: true,
		CreatedAt:  createdAt,
		EditedAt:   editedAt,
	}
	mockSvc.EXPECT().VerifyReceipt(ctx, input).Return(response, nil)
	res, err := h.VerifyReceipt(ctx, &pb.VerifyReceiptRequest{ReceiptId: receiptID.String()})
	assert.NoError(t, err)
	assert.Equal(t, receiptID.String(), res.Id)
	assert.Equal(t, lessonID.String(), *res.LessonId)
	assert.Equal(t, fileID.String(), *res.FileId)
	assert.Equal(t, true, res.IsVerified)
	assert.Equal(t, createdAt, res.CreatedAt.AsTime().Truncate(time.Second))
	assert.Equal(t, editedAt, res.EditedAt.AsTime().Truncate(time.Second))
}

func TestVerifyReceipt_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiptID := uuid.New()
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.VerifyReceiptInput{ReceiptId: receiptID}
	mockSvc.EXPECT().VerifyReceipt(ctx, input).Return(nil, errdefs.ErrNotFound)
	_, err := h.VerifyReceipt(ctx, &pb.VerifyReceiptRequest{ReceiptId: receiptID.String()})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestVerifyReceipt_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiptID := uuid.New()
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.VerifyReceiptInput{ReceiptId: receiptID}
	mockSvc.EXPECT().VerifyReceipt(ctx, input).Return(nil, errdefs.ErrPermissionDenied)
	_, err := h.VerifyReceipt(ctx, &pb.VerifyReceiptRequest{ReceiptId: receiptID.String()})
	assert.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestVerifyReceipt_InvalidReceiptID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	_, err := h.VerifyReceipt(ctx, &pb.VerifyReceiptRequest{ReceiptId: "invalid-uuid"})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetReceiptFile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiptID := uuid.New()
	mockSvc := mocks.NewMockPaymentService(ctrl)
	ctx := context.Background()
	input := &models.GetReceiptFileInput{ReceiptId: receiptID}
	response := &models.ReceiptFileUrl{URL: "http://example.com/receipt.pdf"}
	mockSvc.EXPECT().GetReceiptFile(ctx, input).Return(response, nil)
	res, err := mockSvc.GetReceiptFile(ctx, input)
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/receipt.pdf", res.URL)
}

func TestGetReceiptFile_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiptID := uuid.New()
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.GetReceiptFileInput{ReceiptId: receiptID}
	mockSvc.EXPECT().GetReceiptFile(ctx, input).Return(nil, errdefs.ErrNotFound)
	_, err := h.GetReceiptFile(ctx, &pb.GetReceiptFileRequest{ReceiptId: receiptID.String()})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestGetReceiptFile_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiptID := uuid.New()
	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	input := &models.GetReceiptFileInput{ReceiptId: receiptID}
	mockSvc.EXPECT().GetReceiptFile(ctx, input).Return(nil, errdefs.ErrPermissionDenied)
	_, err := h.GetReceiptFile(ctx, &pb.GetReceiptFileRequest{ReceiptId: receiptID.String()})
	assert.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestGetReceiptFile_InvalidReceiptID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockPaymentService(ctrl)
	h := &PaymentServiceServer{service: mockSvc}
	ctx := context.Background()
	_, err := h.GetReceiptFile(ctx, &pb.GetReceiptFileRequest{ReceiptId: "invalid-uuid"})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}
