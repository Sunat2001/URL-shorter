package authResponse

import (
	"url-shortner/internel/domain/entities/user"
	"url-shortner/internel/lib/api/response"
)

type Response struct {
	response.Response
	User          user.User     `json:"user"`
	AuthTokenInfo AuthTokenInfo `json:"authorisation"`
}

type AuthTokenInfo struct {
	Token string `json:"token"`
	Type  string `json:"type"`
}
