package service

import (
	"context"
	"errors"
	"ne_noy/internal/config"
	"ne_noy/internal/dto"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"
	"strings"

	"github.com/google/uuid"
)

type UserService interface {
	UpdateRole(ctx context.Context, vkId int64, roleUuid uuid.UUID) error
	GetAllUsers(ctx context.Context, fio string) ([]dto.UserMiniDto, error)
	UpdatePermissions(ctx context.Context, permission string, vkId int64, value bool) error
	CreateUser(ctx context.Context, createUserDto dto.CreateUserDto) (*dto.UserDto, error)
	GetUserByVkId(ctx context.Context, vkId int64) (*dto.UserDto, error)
}

type userService struct {
	r  repository.UserRepository
	rr repository.RoleRepository
}

func NewUserService(r repository.UserRepository, rr repository.RoleRepository) UserService {
	return &userService{r: r, rr: rr}
}

func (s *userService) UpdateRole(ctx context.Context, vkId int64, roleUuid uuid.UUID) error {
	updated, err := s.r.Update(ctx, vkId, "role_id", roleUuid)
	if err != nil {
		return err
	}
	if !updated {
		return errors.New("user not found or role not changed")
	}
	return nil
}

func (s *userService) GetAllUsers(ctx context.Context, fio string) ([]dto.UserMiniDto, error) {
	fio = strings.TrimSpace(fio)
	if fio == "" {
		users, err := s.r.GetAll(ctx)
		if err != nil {
			return nil, err
		}
		return s.toMiniDtos(users), nil
	}

	parts := strings.Fields(fio)
	switch len(parts) {
	case 1:
		users, err := s.r.GetAllByFirstNameAndRole(ctx, parts[0])
		if err != nil {
			return nil, err
		}
		return s.toMiniDtos(users), nil

	case 2:
		users, err := s.r.GetAllByFirstNameAndLastNameAndRole(ctx, parts[0], parts[1])
		if err != nil {
			return nil, err
		}
		return s.toMiniDtos(users), nil

	default:
		return nil, errors.New("fio should contain at most first and last name")
	}
}

func (s *userService) UpdatePermissions(ctx context.Context, permission string, vkId int64, value bool) error {
	var field string

	switch permission {
	case "geo":
		field = "geo_available"
	case "notification":
		field = "notification_available"
	default:
		return errors.New("недопустимый параметр permission")
	}

	updated, err := s.r.Update(ctx, vkId, field, value)
	if err != nil {
		return err
	}
	if !updated {
		return errors.New("user not found")
	}
	return nil
}

func (s *userService) CreateUser(ctx context.Context, createUserDto dto.CreateUserDto) (*dto.UserDto, error) {
	defaultRole, err := s.rr.GetByCode(ctx, config.RoleDefault)
	if err != nil {
		return nil, err
	}

	if defaultRole == nil {
		return nil, errors.New("default role not found")
	}

	user := model.User{
		ID:                    uuid.New(),
		FirstName:             createUserDto.FirstName,
		LastName:              createUserDto.LastName,
		VkID:                  createUserDto.VkId,
		PhotoURL:              createUserDto.PhotoURL,
		GeoAvailable:          false,
		NotificationAvailable: false,
		Role:                  defaultRole,
		RoleID:                &defaultRole.ID,
	}

	err = s.r.Create(ctx, &user)
	if err != nil {
		return nil, err
	}

	return s.userModelToDto(user), nil
}

func (s *userService) GetUserByVkId(ctx context.Context, vkId int64) (*dto.UserDto, error) {
	userModel, err := s.r.GetByVkId(ctx, vkId)
	if err != nil {
		return nil, err
	}
	if userModel == nil {
		return nil, nil
	}

	return s.userModelToDto(*userModel), nil
}

func (s *userService) toMiniDtos(users []model.User) []dto.UserMiniDto {
	dtos := make([]dto.UserMiniDto, len(users))
	for i, u := range users {
		dtos[i] = dto.UserMiniDto{
			ID:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			PhotoURL:  u.PhotoURL,
			VkId:      u.VkID,
		}
	}
	return dtos
}

func (s *userService) userModelToDto(user model.User) *dto.UserDto {
	roleDto := dto.RoleDto{
		ID:          user.Role.ID,
		Name:        user.Role.Name,
		DisplayName: user.Role.DisplayName,
	}

	isAdmin := user.Role.Name == config.RoleAdmin ||
		func() bool {
			exists, err := s.r.ExistEventOrg(context.Background(), user.ID) // ← внимание!
			if err != nil {
				return false
			}
			return exists
		}()

	return &dto.UserDto{
		ID:                    &user.ID,
		FirstName:             user.FirstName,
		LastName:              user.LastName,
		GeoAvailable:          user.GeoAvailable,
		NotificationAvailable: user.NotificationAvailable,
		Role:                  roleDto,
		IsAdmin:               isAdmin,
		IsEduParticipant:      user.Role.Name == config.RoleEduPart,
		VkId:                  user.VkID,
		PhotoURL:              user.PhotoURL,
	}
}
