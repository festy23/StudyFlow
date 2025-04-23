package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	paymentpb "paymentservice/pkg/api"
)

type PaymentHandler struct {
	c paymentpb.PaymentServiceClient
}

func NewPaymentHandler(c paymentpb.PaymentServiceClient) *PaymentHandler {
	return &PaymentHandler{c: c}
}

func (h *PaymentHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Group(func(r chi.Router) {
		r.Get("/info/{lesson_id}", h.GetPaymentInfo)
		r.Post("/receipts", h.SubmitReceipt)
		r.Get("/receipts/{id}", h.GetReceipt)
		r.Post("/receipts/{id}/verify", h.VerifyReceipt)
		r.Get("/receipts/{id}/file-url", h.GetReceiptFile)
	})
}

func parseGetPaymentInfo(ctx context.Context, r *http.Request, req *paymentpb.GetPaymentInfoRequest) error {
	id, err := parsePathParam(r, "lesson_id")
	if err != nil {
		return err
	}
	req.LessonId = &id
	return nil
}

func parseGetReceipt(ctx context.Context, r *http.Request, req *paymentpb.GetReceiptRequest) error {
	id, err := parsePathParam(r, "id")
	if err != nil {
		return err
	}
	req.ReceiptId = id
	return nil
}

func parseVerifyReceipt(ctx context.Context, r *http.Request, req *paymentpb.VerifyReceiptRequest) error {
	id, err := parsePathParam(r, "id")
	if err != nil {
		return err
	}
	req.ReceiptId = id
	return nil
}

func parseGetReceiptFile(ctx context.Context, r *http.Request, req *paymentpb.GetReceiptFileRequest) error {
	id, err := parsePathParam(r, "id")
	if err != nil {
		return err
	}
	req.ReceiptId = id
	return nil
}

func (h *PaymentHandler) GetPaymentInfo(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[paymentpb.GetPaymentInfoRequest, paymentpb.PaymentInfo](h.c.GetPaymentInfo, parseGetPaymentInfo, false)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *PaymentHandler) SubmitReceipt(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[paymentpb.SubmitPaymentReceiptRequest, paymentpb.Receipt](h.c.SubmitPaymentReceipt, nil, true)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *PaymentHandler) GetReceipt(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[paymentpb.GetReceiptRequest, paymentpb.Receipt](h.c.GetReceipt, parseGetReceipt, false)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *PaymentHandler) VerifyReceipt(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[paymentpb.VerifyReceiptRequest, paymentpb.Receipt](h.c.VerifyReceipt, parseVerifyReceipt, false)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}

func (h *PaymentHandler) GetReceiptFile(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[paymentpb.GetReceiptFileRequest, paymentpb.ReceiptFileURL](h.c.GetReceiptFile, parseGetReceiptFile, false)
	if err != nil {
		panic(err)
	}
	handler(w, r)
}
