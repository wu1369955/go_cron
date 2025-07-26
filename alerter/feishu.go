package alerter

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

var _ Alerter = &FeishuAlert{}

// NewFeishuAlerter 飞书自定义机器人告警
func NewFeishuAlerter(webhookURL string, signSecret string) *FeishuAlert {
	return &FeishuAlert{
		webhookURL: webhookURL,
		signSecret: signSecret,
	}
}

type FeishuAlert struct {
	webhookURL string
	signSecret string
}

func (a *FeishuAlert) Alert(data map[string]any) {
	var err error
	for range 20 {
		err = a.doAlert(data)
		if err == nil {
			return
		}
	}

	slog.Error("feishu alert failed", "error", err)
}

func (a *FeishuAlert) doAlert(data map[string]any) error {
	text, err := json.Marshal(data)
	if err != nil {
		return err
	}

	timestamp := time.Now().Unix()
	sign, err := genSign(a.signSecret, timestamp)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(map[string]any{
		"timestamp": strconv.FormatInt(timestamp, 10),
		"msg_type":  "text",
		"sign":      sign,
		"content": map[string]any{
			"text": string(text),
		},
	})
	if err != nil {
		return err
	}

	// 发送请求
	resp, err := http.Post(a.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Request failed with status code: %d\n", resp.StatusCode)
	}

	var result feishuResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode resp body: %w", err)
	}
	if result.Code != 0 {
		return fmt.Errorf("code=%v, msg=%v", result.Code, result.Msg)
	}

	return nil
}

type feishuResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func genSign(secret string, timestamp int64) (string, error) {
	// timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret

	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}
