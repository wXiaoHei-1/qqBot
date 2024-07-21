package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	constant "qqbot/constant"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// SendMessage 向gpt发送消息
func SendMessage(context string, dashScopeAPIKey string) string {
	// 创建请求体
	requestBody := RequestBody{
		Model: constant.DashScopeModel,
		Messages: []Message{
			{
				Role:    "user",
				Content: context,
			},
		},
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error serializing request body:", err)
		return ""
	}
	req, err := http.NewRequest("POST", constant.DashScopeAPIURL, bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return ""
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", dashScopeAPIKey))
	req.Header.Set("Content-Type", "application/json")

	// 发送 HTTP 请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return ""
	}
	defer resp.Body.Close()

	// 读取 HTTP 响应体
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading HTTP response body:", err)
		return ""
	}

	// 解析 HTTP 响应体
	var responseBody map[string]interface{}
	err = json.Unmarshal(respBody, &responseBody)
	if err != nil {
		fmt.Println("Error parsing HTTP response body:", err)
		return ""
	}
	return responseBody["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
}
