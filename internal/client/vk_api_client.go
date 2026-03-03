package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"ne_noy/internal/dto"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type VkApiClient interface {
	GetVkUsers(userIds []string) ([]dto.CreateUserDto, error)
	SendNotification(userIds []string, messageText, fragment string) (dto.SendMessageResponse, error)
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

func (v vkApiClient) SendNotification(userIds []string, messageText, fragment string) (dto.SendMessageResponse, error) {
	dtoJson, _ := json.Marshal(dto.SendMessageDto{
		AccessToken: v.serviceKey,
		UserIds:     mapUsers(userIds),
		Message:     messageText,
		Fragment:    &fragment,
	})

	reader := bytes.NewReader(dtoJson)
	client := http.Client{Timeout: time.Duration(10) * time.Second}
	res, err := client.Post(v.baseUrl+"/notifications.sendMessage", "application/json", reader)
	if err != nil {
		return dto.SendMessageResponse{}, err
	}
	defer res.Body.Close()
	var response dto.SendMessageResponse
	var b []byte
	_, err = res.Body.Read(b)
	if err != nil {
		return dto.SendMessageResponse{}, err
	}
	err = json.Unmarshal(b, &response)
	if err != nil {
		return dto.SendMessageResponse{}, err
	}
	return response, nil
}

func mapUsers(userIds []string) string {
	result := strings.Builder{}
	for i, userId := range userIds {
		result.WriteString(userId)
		if i != len(userIds)-1 {
			result.WriteString(",")
		}
	}
	return result.String()
}

func NewVkApiClient(serviceKey, baseUrl string) VkApiClient {
	return &vkApiClient{serviceKey: serviceKey, baseUrl: baseUrl}
}
