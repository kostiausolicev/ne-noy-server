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
	CreateUser(user model.User) (*model.User, error)
	GetUserByVkId(vkId int64) (*dto.UserDto, error)
}

type userService struct {
	r repository.UserRepository
}

func (u userService) CreateUser(user model.User) (*model.User, error) {
	defaultRole, err := u.r.GetRole()
	if err != nil {
		return nil, err
	}
	user.Role = defaultRole
	newUser, err := u.r.Create(&user)
	if err != nil {
		return nil, err
	}
	return newUser, nil
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
