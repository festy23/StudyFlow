package service

import (
	"common_library/ctxdata"
	"context"
	"github.com/google/uuid"
	"strings"
	"userservice/internal/authorization"
	"userservice/internal/errdefs"
	"userservice/internal/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, input *model.RepositoryCreateUserInput) (*model.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input *model.UpdateUserInput) (*model.User, error)

	CreateTutorProfile(ctx context.Context, input *model.RepositoryCreateTutorProfileInput) (*model.TutorProfile, error)
	GetTutorProfile(ctx context.Context, userId uuid.UUID) (*model.TutorProfile, error)
	UpdateTutorProfile(ctx context.Context, userId uuid.UUID, input *model.UpdateTutorProfileInput) (*model.TutorProfile, error)

	CreateTelegramAccount(ctx context.Context, input *model.RepositoryCreateTelegramAccountInput) (*model.TelegramAccount, error)
	GetTelegramAccount(ctx context.Context, userId uuid.UUID) (*model.TelegramAccount, error)
	GetTelegramAccountByTelegramId(ctx context.Context, telegramId int64) (*model.TelegramAccount, error)
	ExistsByTelegramID(ctx context.Context, telegramID int64) (bool, error)
}

type TutorStudentsRepository interface {
}

type UserService struct {
	userRepository     UserRepository
	tsRepository       TutorStudentsRepository
	telegramAuthSecret string
}

func NewUserService(
	userRepository UserRepository,
	tutorStudentsRepository TutorStudentsRepository,
	telegramAuthSecret string,
) *UserService {
	return &UserService{userRepository, tutorStudentsRepository, telegramAuthSecret}
}

func (s *UserService) RegisterViaTelegram(ctx context.Context, input *model.RegisterViaTelegramInput) (*model.User, error) {
	if !input.Role.IsValid() {
		return nil, errdefs.ValidationErr
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	userInput := &model.RepositoryCreateUserInput{
		Id:           id,
		Role:         input.Role,
		AuthProvider: model.AuthProviderTelegram,
		Status:       model.UserStatusActive,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Timezone:     input.Timezone,
	}

	user, err := s.userRepository.CreateUser(ctx, userInput)
	if err != nil {
		return nil, err
	}

	id, err = uuid.NewV7()
	if err != nil {
		return nil, err
	}

	tgAccountInput := &model.RepositoryCreateTelegramAccountInput{
		Id:         id,
		UserId:     user.Id,
		TelegramId: input.TelegramId,
		Username:   input.Username,
	}

	_, err = s.userRepository.CreateTelegramAccount(ctx, tgAccountInput)
	if err != nil {
		return nil, err
	}

	if user.Role == model.RoleTutor {
		id, err = uuid.NewV7()
		if err != nil {
			return nil, err
		}
		tutorProfileInput := &model.RepositoryCreateTutorProfileInput{
			Id:     id,
			UserId: user.Id,
		}

		_, err := s.userRepository.CreateTutorProfile(ctx, tutorProfileInput)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}

func (s *UserService) Authorize(ctx context.Context, input *model.AuthorizeInput) (*model.User, error) {
	header := input.AuthorizationHeader
	if strings.HasPrefix(header, "telegram") {
		return s.authorizeWithTelegram(ctx, strings.Trim(strings.TrimPrefix(header, "telegram"), " "))
	}

	return nil, errdefs.AuthenticationErr
}

func (s *UserService) authorizeWithTelegram(ctx context.Context, header string) (*model.User, error) {
	telegramId, err := authorization.GetTelegramId(s.telegramAuthSecret, header)
	if err != nil {
		return nil, err
	}

	tgAccount, err := s.userRepository.GetTelegramAccountByTelegramId(ctx, telegramId)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepository.GetUser(ctx, tgAccount.UserId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetMe(ctx context.Context) (*model.User, error) {
	userId, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, errdefs.AuthenticationErr
	}

	id, err := uuid.Parse(userId)
	if err != nil {
		return nil, errdefs.AuthenticationErr
	}

	user, err := s.userRepository.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserPublic(ctx context.Context, id uuid.UUID) (*model.UserPublic, error) {
	user, err := s.userRepository.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := &model.UserPublic{
		Id:        user.Id,
		Role:      user.Role,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
	return resp, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, input *model.UpdateUserInput) (*model.User, error) {
	user, err := s.userRepository.UpdateUser(ctx, id, input)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetTutorProfile(ctx context.Context, userId uuid.UUID) (*model.TutorProfile, error) {
	return nil, nil
}
