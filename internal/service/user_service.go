package service

import (
	"errors"
	"ne_noy/internal/config"
	"ne_noy/internal/dto"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"

	"github.com/google/uuid"
)

func NewUserService(ur repository.UserRepository) UserService {
	return userService{r: ur}
}

type UserService interface {
	UpdatePermissions(permission string, vkId int64, value bool) error
	CreateUser(user dto.UserDto) (*dto.UserDto, error)
	GetUserByVkId(vkId int64) (*dto.UserDto, error)
}

type userService struct {
	r repository.UserRepository
}

func (u userService) UpdatePermissions(permission string, vkId int64, value bool) error {
	switch permission {
	case "geo":
		u.r.Update(vkId, "geo_available", value)
	case "notification":
		u.r.Update(vkId, "notification_available", value)
	default:
		return errors.New("недопустимый параметр")
	}
	return nil
}

func (u userService) CreateUser(createUserDto dto.UserDto) (*dto.UserDto, error) {
	defaultRole, err := u.r.GetRole()
	if err != nil {
		return nil, err
	}
	user := model.User{
		FirstName:             createUserDto.FirstName,
		LastName:              createUserDto.LastName,
		VkID:                  createUserDto.VkId,
		PhotoURL:              createUserDto.PhotoURL,
		GeoAvailable:          createUserDto.GeoAvailable,
		NotificationAvailable: createUserDto.NotificationAvailable,
		Role:                  defaultRole,
		RoleID:                &defaultRole.ID,
	}
	newUser, err := u.r.Create(&user)
	if err != nil {
		return nil, err
	}
	return u.userModelToDto(newUser), nil
}

func (u userService) GetUserByVkId(vkId int64) (*dto.UserDto, error) {
	userModel, err := u.r.GetByVkId(vkId)
	if err != nil {
		return nil, err
	}
	if userModel == nil {
		return nil, nil
	}
	return u.userModelToDto(userModel), nil
}

func (u userService) userModelToDto(userModel *model.User) *dto.UserDto {
	userDto := &dto.UserDto{}
	roleDto := &dto.RoleDto{}
	roleDto.ID = userModel.Role.ID
	roleDto.Name = userModel.Role.Name
	roleDto.DisplayName = userModel.Role.DisplayName

	userDto.ID = userModel.ID
	userDto.FirstName = userModel.FirstName
	userDto.LastName = userModel.LastName
	userDto.GeoAvailable = userModel.GeoAvailable
	userDto.NotificationAvailable = userModel.NotificationAvailable
	userDto.Role = *roleDto
	userDto.IsAdmin = userModel.Role.Name == config.RoleAdmin || func(userId uuid.UUID) bool {
		e, err := u.r.ExistEventOrg(userId)
		if err != nil {
			return false
		}
		return e
	}(userModel.ID)
	userDto.IsEduParticipant = userModel.Role.Name == config.RoleEduPart
	userDto.VkId = userModel.VkID
	userDto.PhotoURL = userModel.PhotoURL

	return userDto
}
