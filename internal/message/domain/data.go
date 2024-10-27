package domain

type Data struct {
	ChatId               int64 `json:"ChatId"`
	Profit               int   `json:"Profit"`
	FailedUpdateAttempts int   `json:"FailedUpdateAttempts"`
}
