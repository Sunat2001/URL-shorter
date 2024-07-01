package redirectInfo

type RedirectInfo struct {
	Id          int64  `json:"id,omitempty"`
	Ip          string `json:"ip"`
	Os          string `json:"os"`
	Platform    string `json:"platform"`
	Browser     string `json:"browser"`
	Created     string `json:"created"`
	Country     string `json:"country,omitempty"`
	City        string `json:"city,omitempty"`
	CountryCode string `json:"countryCode,omitempty"`
}
