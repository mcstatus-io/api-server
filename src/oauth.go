package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type DiscordAccessExchange struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

type DiscordUser struct {
	Email string `json:"email"`
}

type GitHubAccessExchange struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

type GitHubEmail struct {
	Email   string `json:"email"`
	Primary bool   `json:"primary"`
}

func ExchangeDiscordAccessToken(code string) (*DiscordAccessExchange, error) {
	requestBody := &url.Values{}
	requestBody.Set("grant_type", "authorization_code")
	requestBody.Set("code", code)
	requestBody.Set("redirect_uri", config.Discord.RedirectURI)

	req, err := http.NewRequest("POST", "https://discord.com/api/v10/oauth2/token", strings.NewReader(requestBody.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", config.Discord.ClientID, config.Discord.Secret)))))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord: unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var response DiscordAccessExchange

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func GetDiscordUser(accessToken string) (*DiscordUser, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/users/@me", nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord: unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var response DiscordUser

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func ExchangeGitHubAccessToken(code string) (*GitHubAccessExchange, error) {
	requestBody := &url.Values{}
	requestBody.Set("client_id", config.GitHub.ClientID)
	requestBody.Set("client_secret", config.GitHub.Secret)
	requestBody.Set("code", code)
	requestBody.Set("redirect_uri", config.GitHub.RedirectURI)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(requestBody.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github: unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var response GitHubAccessExchange

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func GetGitHubEmails(accessToken string) ([]*GitHubEmail, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord: unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	response := make([]*GitHubEmail, 0)

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, err
	}

	return response, nil
}
