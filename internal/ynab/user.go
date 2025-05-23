package ynab

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) CheckConnection() (bool, error) {
	url := fmt.Sprintf("%s/user", root)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	userResponse := struct {
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	if userResponse.Data.User.ID == "" {
		return false, fmt.Errorf("user ID is empty")
	}

	return true, nil
}
