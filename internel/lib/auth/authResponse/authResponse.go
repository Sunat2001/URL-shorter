package authResponse

import (
	"url-shortner/internel/lib/api/response"
	"url-shortner/internel/storage/sqlite"
)

type Response struct {
	response.Response
	User          sqlite.User   `json:"user"`
	AuthTokenInfo AuthTokenInfo `json:"authorisation"`
}

type AuthTokenInfo struct {
	Token string `json:"token"`
	Type  string `json:"type"`
}
