package handler

import (
	"common_library/logging"
	"context"
	filepb "fileservice/pkg/api"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

type FileHandler struct {
	c filepb.FileServiceClient
}

func NewFileHandler(c filepb.FileServiceClient) *FileHandler {
	return &FileHandler{c: c}
}

func (h *FileHandler) RegisterRoutes(r chi.Router) {
	r.Post("/init-upload", h.InitUpload)
	r.Get("/{id}/meta", h.GetFileMeta)
}

func (h *FileHandler) InitUpload(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[filepb.InitUploadRequest, filepb.InitUploadResponse](h.c.InitUpload, nil, true)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *FileHandler) GetFileMeta(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[filepb.GetFileMetaRequest, filepb.File](h.c.GetFileMeta, getFileMetaParsePath, false)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func getFileMetaParsePath(ctx context.Context, httpReq *http.Request, grpcReq *filepb.GetFileMetaRequest) error {
	id := chi.URLParam(httpReq, "id")
	if id == "" {
		return fmt.Errorf("%w: %s", BadRequestError, "studentId is required")
	}
	grpcReq.FileId = id

	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Debug(ctx, "file id added to request", zap.Any("req", grpcReq))
	}
	return nil
}
