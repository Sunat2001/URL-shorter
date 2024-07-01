package urlInfo

import "url-shortner/internel/domain/entities/user"

type UrlInfo struct {
	Alias string    `json:"alias"`
	Url   string    `json:"url"`
	User  user.User `json:"user"`
}
