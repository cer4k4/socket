package domain

type Entity struct {
	EntityType int        `json:"EntityType"`
	Operation  int        `json:"Operation"`
	Data       EntityData `json:"Data"`
}

type EntityData struct {
	ChatId               int64 `json:"ChatId"`
	Profit               int   `json:"Profit"`
	FailedUpdateAttempts int   `json:"FailedUpdateAttempts"`
}

type Data struct {
	ChatId               int64 `json:"ChatId"`
	Profit               int   `json:"Profit"`
	FailedUpdateAttempts int   `json:"FailedUpdateAttempts"`
}

type Message struct {
	EntityType int    `json:"EntityType"`
	Operation  int    `json:"Operation"`
	Data       string `json:"Data"`
}
