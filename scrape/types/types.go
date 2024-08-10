package types

type Site struct {
	Query   string `json:"query"`
	Page    string `json:"page"`
	MaxPage string `json:"maxPage"`
}

type MainSitesStrct struct {
	Data  SitesStrct `json:"data"`
	Sites string     `json:"sites"`
}
type SitesStrct map[string]Site

type JobStrct struct {
	Title       string
	Company     string
	Date        string
	Salary      string
	Details     string
	JobLink     string
	CompanyRate string
	Place       string
	Modality    string
	Location    string
	ApplyVia    string
	Site        string
	Saved       string
}
