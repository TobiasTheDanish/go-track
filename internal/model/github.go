package model

type AuthUserRes struct {
	AccessToken           string `json:"access_token"`
	ExpiresIn             int    `json:"expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
	Scope                 string `json:"scope"`
	TokenType             string `json:"token_type"`
}

type AuthorizedUser struct {
	Username string `json:"login"`
	OrgUrl   string `json:"organizations_url"`
}
