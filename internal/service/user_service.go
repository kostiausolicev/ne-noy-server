package service

import (
	"ne_noy/internal/dto"
	"ne_noy/internal/repository"
)

func NewUserService(ur repository.UserRepository) UserService {
	return userService{r: ur}
}

type UserService interface {
	GetUserByVkId(vkId int64) (*dto.UserDto, error)
}

type userService struct {
	r repository.UserRepository
}

func (u userService) GetUserByVkId(vkId int64) (*dto.UserDto, error) {
	userDto := &dto.UserDto{}
	userModel, err := u.r.GetByVkId(vkId)
	if err != nil {
		return nil, err
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

	return userDto, nil
}
