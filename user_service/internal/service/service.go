package service

import (
	"common_library/ctxdata"
	"common_library/logging"
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"slices"
	"strings"
	"userservice/internal/authorization"
	"userservice/internal/errdefs"
	"userservice/internal/model"
)

type UserRepository interface {
	NewUserCreationRepositoryTx(ctx context.Context) (UserCreationRepositoryTx, error)

	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input *model.UpdateUserInput) (*model.User, error)

	GetTutorProfile(ctx context.Context, userId uuid.UUID) (*model.TutorProfile, error)
	UpdateTutorProfile(ctx context.Context, userId uuid.UUID, input *model.UpdateTutorProfileInput) (*model.TutorProfile, error)

	GetTelegramAccount(ctx context.Context, userId uuid.UUID) (*model.TelegramAccount, error)
	GetTelegramAccountByTelegramId(ctx context.Context, telegramId int64) (*model.TelegramAccount, error)
}

type UserCreationRepositoryTx interface {
	CreateUser(ctx context.Context, input *model.RepositoryCreateUserInput) (*model.User, error)
	CreateTutorProfile(ctx context.Context, input *model.RepositoryCreateTutorProfileInput) (*model.TutorProfile, error)
	CreateTelegramAccount(ctx context.Context, input *model.RepositoryCreateTelegramAccountInput) (*model.TelegramAccount, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type TutorStudentsRepository interface {
	CreateTutorStudent(ctx context.Context, input *model.RepositoryCreateTutorStudentInput) (*model.TutorStudent, error)
	UpdateTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID, input *model.UpdateTutorStudentInput) (*model.TutorStudent, error)
	GetTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) (*model.TutorStudent, error)
	DeleteTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) error
	// ListTutorStudents Set tutorId or studentId to UUID.Nil to search by one parameter
	ListTutorStudents(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) ([]*model.TutorStudent, error)
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

	repo, err := s.userRepository.NewUserCreationRepositoryTx(ctx)
	if err != nil {
		return nil, err
	}

	defer func(repo UserCreationRepositoryTx, ctx context.Context) {
		err := repo.Rollback(ctx)
		if err != nil {
			logger, ok := logging.GetFromContext(ctx)
			if ok {
				logger.Error(ctx, fmt.Sprintf("Failed to Rollback"), zap.Error(err))
			}
		}
	}(repo, ctx)

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

	user, err := repo.CreateUser(ctx, userInput)
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

	_, err = repo.CreateTelegramAccount(ctx, tgAccountInput)
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

		_, err := repo.CreateTutorProfile(ctx, tutorProfileInput)
		if err != nil {
			return nil, err
		}
	}

	err = repo.Commit(ctx)
	if err != nil {
		return nil, err
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
	if err := ensureCurrentUserIs(ctx, id); err != nil {
		return nil, err
	}

	user, err := s.userRepository.UpdateUser(ctx, id, input)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetTutorProfile(ctx context.Context, userId uuid.UUID) (*model.TutorProfile, error) {
	if err := ensureCurrentUserIs(ctx, userId); err != nil {
		return nil, err
	}

	profile, err := s.userRepository.GetTutorProfile(ctx, userId)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *UserService) UpdateTutorProfile(ctx context.Context, userId uuid.UUID, input *model.UpdateTutorProfileInput) (*model.TutorProfile, error) {
	if err := ensureCurrentUserIs(ctx, userId); err != nil {
		return nil, err
	}

	profile, err := s.userRepository.UpdateTutorProfile(ctx, userId, input)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *UserService) CreateTutorStudent(ctx context.Context, input *model.CreateTutorStudentInput) (*model.TutorStudent, error) {
	if err := ensureCurrentUserIs(ctx, input.TutorId); err != nil {
		return nil, err
	}

	if err := ensureCurrentUserRole(ctx, model.RoleTutor); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	tsInput := &model.RepositoryCreateTutorStudentInput{
		Id:                   id,
		TutorId:              input.TutorId,
		StudentId:            input.StudentId,
		LessonPriceRub:       input.LessonPriceRub,
		LessonConnectionLink: input.LessonConnectionLink,
		Status:               model.TutorStudentStatusInvited,
	}
	ts, err := s.tsRepository.CreateTutorStudent(ctx, tsInput)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

func (s *UserService) GetTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) (*model.TutorStudent, error) {
	if err := ensureCurrentUserIs(ctx, tutorId, studentId); err != nil {
		return nil, err
	}
	ts, err := s.tsRepository.GetTutorStudent(ctx, tutorId, studentId)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

func (s *UserService) UpdateTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID, input *model.UpdateTutorStudentInput) (*model.TutorStudent, error) {
	if err := ensureCurrentUserIs(ctx, tutorId); err != nil {
		return nil, err
	}
	ts, err := s.tsRepository.UpdateTutorStudent(ctx, tutorId, studentId, input)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

func (s *UserService) DeleteTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) error {
	if err := ensureCurrentUserIs(ctx, tutorId); err != nil {
		return err
	}

	if err := s.tsRepository.DeleteTutorStudent(ctx, tutorId, studentId); err != nil {
		return err
	}

	return nil
}

func (s *UserService) ListTutorStudents(ctx context.Context, tutorId uuid.UUID) ([]*model.TutorStudent, error) {
	if err := ensureCurrentUserIs(ctx, tutorId); err != nil {
		return nil, err
	}

	resp, err := s.tsRepository.ListTutorStudents(ctx, tutorId, uuid.Nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *UserService) ListTutorStudentsForStudent(ctx context.Context, studentId uuid.UUID) ([]*model.TutorStudent, error) {
	if err := ensureCurrentUserIs(ctx, studentId); err != nil {
		return nil, err
	}

	resp, err := s.tsRepository.ListTutorStudents(ctx, uuid.Nil, studentId)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *UserService) ResolveTutorStudentContext(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) (*model.TutorStudentContext, error) {
	if err := ensureCurrentUserIs(ctx, tutorId, studentId); err != nil {
		return nil, err
	}

	tutorProfile, err := s.userRepository.GetTutorProfile(ctx, tutorId)
	if err != nil {
		return nil, err
	}

	ts, err := s.tsRepository.GetTutorStudent(ctx, tutorId, studentId)
	if err != nil {
		return nil, err
	}

	resp := &model.TutorStudentContext{
		RelationshipStatus:   ts.Status,
		LessonPriceRub:       tutorProfile.LessonPriceRub,
		LessonConnectionLink: tutorProfile.LessonConnectionLink,
		PaymentInfo:          tutorProfile.PaymentInfo,
	}

	if ts.LessonPriceRub != nil {
		resp.LessonPriceRub = ts.LessonPriceRub
	}

	if ts.LessonConnectionLink != nil {
		resp.LessonConnectionLink = ts.LessonConnectionLink
	}

	return resp, nil
}

func (s *UserService) AcceptInvitationFromTutor(ctx context.Context, tutorId uuid.UUID) error {
	id, err := getUserId(ctx)
	if err != nil {
		return err
	}

	if err := ensureCurrentUserRole(ctx, model.RoleStudent); err != nil {
		return err
	}

	status := model.TutorStudentStatusActive
	_, err = s.tsRepository.UpdateTutorStudent(ctx, tutorId, id, &model.UpdateTutorStudentInput{Status: &status})
	if err != nil {
		return err
	}

	return nil
}

func getUserId(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return uuid.Nil, errdefs.AuthenticationErr
	}

	idUUID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errdefs.AuthenticationErr
	}

	return idUUID, nil
}

func getRole(ctx context.Context) (model.Role, error) {
	roleString, ok := ctxdata.GetUserRole(ctx)
	if !ok {
		return "", errdefs.AuthenticationErr
	}
	role := model.Role(roleString)
	if !role.IsValid() {
		return "", errdefs.AuthenticationErr
	}

	return role, nil
}

func ensureCurrentUserIs(ctx context.Context, ids ...uuid.UUID) error {
	currentUserId, err := getUserId(ctx)
	if err != nil {
		return err
	}
	if !slices.Contains(ids, currentUserId) {
		return errdefs.ErrPermissionDenied
	}
	return nil
}

func ensureCurrentUserRole(ctx context.Context, role model.Role) error {
	userRole, err := getRole(ctx)
	if err != nil {
		return err
	}
	if userRole != role {
		return errdefs.ErrPermissionDenied
	}
	return nil
}
