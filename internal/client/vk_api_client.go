package client

import (
	"context"
	"encoding/json"
	"fmt"
	"ne_noy/internal/dto"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type VkApiClient interface {
	GetVkUsers(userIds []string) ([]dto.CreateUserDto, error)
	SendNotificationForUsers(userIds []int64, messageText, fragment string) (dto.SendMessageResponse, error)
	CreateChat(ctx context.Context, name string, userIds []uuid.UUID) (int64, error)
	AddUserToChat(ctx context.Context, token string, chatId, userVkId int64) error
}

type vkApiClient struct {
	serviceKey string
	baseUrl    string
}

func (v vkApiClient) GetVkUsers(userIds []string) ([]dto.CreateUserDto, error) {
	// Создаем form data
	formData := url.Values{}
	formData.Set("access_token", v.serviceKey)
	formData.Set("user_ids", strings.Join(userIds, ","))
	formData.Set("fields", "photo_max")
	formData.Set("v", "5.199") // TODO в конфиги

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.PostForm(v.baseUrl+"/users.get", formData)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var response dto.UsersGetResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	users := make([]dto.CreateUserDto, len(response.Users))

	for i, userInfo := range response.Users {
		if userInfo.FirstName == "DELETED" {
			continue
		}
		users[i] = dto.CreateUserDto{
			VkId:      userInfo.VkId,
			FirstName: userInfo.FirstName,
			LastName:  userInfo.LastName,
			PhotoURL:  userInfo.PhotoId,
		}
	}

	return users, nil
}

func (v vkApiClient) SendNotificationForUsers(userIds []int64, messageText, fragment string) (dto.SendMessageResponse, error) {
	formData := url.Values{}
	formData.Set("access_token", v.serviceKey)
	formData.Set("user_ids", mapUsers(userIds))
	formData.Set("message", messageText)
	formData.Set("fragment", fragment)
	formData.Set("v", "5.199")

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.PostForm(v.baseUrl+"/notifications.sendMessage", formData)
	if err != nil {
		return dto.SendMessageResponse{}, err
	}
	defer res.Body.Close()

	var response dto.SendMessageResponse
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return dto.SendMessageResponse{}, err
	}
	return response, nil
}

func (v vkApiClient) CreateChat(ctx context.Context, name string, userIds []uuid.UUID) (int64, error) {
	return 0, nil
}

func (v vkApiClient) AddUserToChat(ctx context.Context, token string, chatId, userVkId int64) error {
	return nil
}

func mapUsers(userIds []int64) string {
	result := strings.Builder{}
	for i, userId := range userIds {
		result.WriteString(strconv.FormatInt(userId, 10))
		if i != len(userIds)-1 {
			result.WriteString(",")
		}
	}
	return result.String()
}

func NewVkApiClient(serviceKey, baseUrl string) VkApiClient {
	return &vkApiClient{serviceKey: serviceKey, baseUrl: baseUrl}
}
