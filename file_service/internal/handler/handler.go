package handler

import (
	"context"
	"fileservice/internal/model"
	pb "fileservice/pkg/api"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FileService interface {
	InitUpload(ctx context.Context, input *model.InitUploadInput) (*model.InitUpload, error)
	GenerateDownloadURL(ctx context.Context, fileId uuid.UUID) (string, error)
	GetFileMeta(ctx context.Context, fileId uuid.UUID) (*model.File, error)
}

type FileHandler struct {
	pb.UnimplementedFileServiceServer
	fileService FileService
}

func NewFileHandler(fileService FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

func (h *FileHandler) InitUpload(ctx context.Context, req *pb.InitUploadRequest) (*pb.InitUploadResponse, error) {
	id, err := uuid.Parse(req.UploadedBy)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	input := &model.InitUploadInput{
		UploadedBy: id,
		Filename:   req.Filename,
	}

	resp, err := h.fileService.InitUpload(ctx, input)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toPbInitUpload(resp), nil
}

func (h *FileHandler) GenerateDownloadURL(ctx context.Context, req *pb.GenerateDownloadURLRequest) (*pb.DownloadURL, error) {
	id, err := uuid.Parse(req.FileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := h.fileService.GenerateDownloadURL(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DownloadURL{Url: resp}, nil
}

func (h *FileHandler) GetFileMeta(ctx context.Context, req *pb.GetFileMetaRequest) (*pb.File, error) {
	id, err := uuid.Parse(req.FileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := h.fileService.GetFileMeta(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toPbFile(resp), nil
}

func toPbInitUpload(init *model.InitUpload) *pb.InitUploadResponse {
	return &pb.InitUploadResponse{
		FileId:    init.FileId.String(),
		UploadUrl: init.UploadURL,
		Method:    init.Method,
	}
}

func toPbFile(file *model.File) *pb.File {
	return &pb.File{
		Id:         file.Id.String(),
		UploadedBy: file.UploadedBy.String(),
		Filename:   file.Filename,
		CreatedAt:  timestamppb.New(file.CreatedAt),
	}
}
