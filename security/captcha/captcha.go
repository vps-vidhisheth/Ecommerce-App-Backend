package security

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
)

type RecaptchaResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func VerifyRecaptcha(token string) error {
	recaptchaSecret := viper.GetString("RECAPTCHA_SECRET")
	if recaptchaSecret == "" {
		return fmt.Errorf("RECAPTCHA_SECRET not set in environment variables")
	}

	form := url.Values{}
	form.Set("secret", recaptchaSecret)
	form.Set("response", token)

	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", form)
	if err != nil {
		return fmt.Errorf("failed to verify recaptcha: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result RecaptchaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse recaptcha response: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("captcha verification failed")
	}

	return nil
}
