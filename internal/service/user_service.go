package service

import (
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
	CreateUser(user dto.UserDto) (*dto.UserDto, error)
	GetUserByVkId(vkId int64) (*dto.UserDto, error)
}

type userService struct {
	r repository.UserRepository
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
	userDto := &dto.UserDto{}

	roleDto := &dto.RoleDto{}
	roleDto.ID = newUser.Role.ID
	roleDto.Name = newUser.Role.Name
	roleDto.DisplayName = newUser.Role.DisplayName

	userDto.ID = newUser.ID
	userDto.FirstName = newUser.FirstName
	userDto.LastName = newUser.LastName
	userDto.GeoAvailable = newUser.GeoAvailable
	userDto.NotificationAvailable = newUser.NotificationAvailable
	userDto.Role = *roleDto
	userDto.IsAdmin = newUser.Role.Name == config.RoleAdmin || func(userId uuid.UUID) bool {
		e, err := u.r.ExistEventOrg(userId)
		if err != nil {
			return false
		}
		return e
	}(newUser.ID)
	userDto.IsEduParticipant = newUser.Role.Name == config.RoleEduPart
	userDto.VkId = newUser.VkID
	userDto.PhotoURL = newUser.PhotoURL

	return userDto, nil
}

func (u userService) GetUserByVkId(vkId int64) (*dto.UserDto, error) {
	userDto := &dto.UserDto{}
	userModel, err := u.r.GetByVkId(vkId)
	if err != nil {
		return nil, err
	}
	if userModel == nil {
		return nil, nil
	}
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

	return userDto, nil
}
