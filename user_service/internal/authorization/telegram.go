package authorization

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"
	"userservice/internal/errdefs"
)

func GetTelegramId(secret string, header string) (int64, error) {
	payload := strings.Split(header, ":")
	if len(payload) != 3 {
		return 0, errdefs.AuthenticationErr
	}
	tgId, err := strconv.ParseInt(payload[2], 10, 64)
	if err != nil {
		return 0, errdefs.AuthenticationErr
	}
	timestamp, err := strconv.ParseInt(payload[1], 10, 64)
	if err != nil {
		return 0, errdefs.AuthenticationErr
	}
	now := time.Now().Unix()
	var diffSeconds int64 = 5 * 60
	if !(now-diffSeconds < timestamp && timestamp < now+diffSeconds) {
		return 0, errdefs.AuthenticationErr
	}

	message := fmt.Sprintf("%s:%s", payload[0], payload[1])
	if !ValidMAC(message, secret, payload[2]) {
		return 0, errdefs.AuthenticationErr
	}

	return tgId, nil
}

func ValidMAC(message, key, messageMAC string) bool {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)
	return hmac.Equal([]byte(messageMAC), expectedMAC)
}
