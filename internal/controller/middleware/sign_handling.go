package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"ne_noy/internal/config"
	"ne_noy/internal/controller"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	tokenPartSeparator       = "."
	payloadPartSeparator     = ";"
	payloadKeyValueSeparator = "="

	signParamAppID     = "app_id"
	signParamRequestID = "request_id"
	signParamTimestamp = "ts"
	signParamUserID    = "user_id"
)

func AuthMiddleware(secret string, appId int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(controller.HeaderAuthorization)
		params, err := validateToken(header, secret, appId)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		role := params[config.UserRoleContextKey]
		vkId := params[config.UserVkIdContextKey]

		c.Set(config.UserRoleContextKey, role)
		c.Set(config.UserVkIdContextKey, vkId)

		c.Next()
	}
}

func validateToken(token, secret string, appId int64) (map[string]interface{}, error) {
	payload, ts, sign, err := separateToken(token)
	if err != nil {
		return nil, err
	}
	expectedSign := generateSign(payload, secret, ts, appId)
	if expectedSign != sign {
		return nil, controller.AuthorizationError
	}
	params := make(map[string]interface{})
	params[config.UserRoleContextKey] = findKeyInPayload(payload, config.UserRoleContextKey)
	params[config.UserVkIdContextKey], _ = strconv.ParseInt(findKeyInPayload(payload, config.UserVkIdContextKey), 10, 64)
	return params, nil
}

func separateToken(token string) (payload string, ts int64, sign string, err error) {
	arr := strings.Split(token, tokenPartSeparator)
	if len(arr) != 3 {
		err = controller.InvalidAuthTokenError
		return
	}
	payload = arr[0]
	ts, err = strconv.ParseInt(arr[1], 10, 64)
	if err != nil {
		return
	}
	sign = arr[2]
	return
}

func generateSign(payload, secret string, ts, appId int64) string {
	// Извлекаем user_id из payload (предполагается формат "user_id=123;other_data")
	var userId int64
	userId, _ = strconv.ParseInt(findKeyInPayload(payload, config.UserVkIdContextKey), 10, 64)
	hashParams := map[string]interface{}{
		signParamAppID:     appId,
		signParamRequestID: payload,
		signParamTimestamp: ts,
		signParamUserID:    userId,
	}

	// Сортируем параметры по ключам
	keys := make([]string, 0, len(hashParams))
	for k := range hashParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Формируем строку параметров
	values := url.Values{}
	for _, k := range keys {
		switch v := hashParams[k].(type) {
		case string:
			values.Add(k, v)
		case int:
			values.Add(k, strconv.Itoa(v))
		case int64:
			values.Add(k, strconv.FormatInt(v, 10))
		}
	}
	signParamsQuery := values.Encode()

	// Вычисляем HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signParamsQuery))
	hash := mac.Sum(nil)

	signature := base64.RawURLEncoding.EncodeToString(hash)
	return signature
}

func findKeyInPayload(payload, key string) string {
	var value string
	parts := strings.Split(payload, payloadPartSeparator)
	for _, part := range parts {
		prefix := key + payloadKeyValueSeparator
		if strings.HasPrefix(part, prefix) {
			value = strings.TrimPrefix(part, prefix)
			break
		}
	}
	return value
}
