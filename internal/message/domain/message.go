package domain

type Message struct {
	EntityType int    `json:"EntityType"`
	Operation  int    `json:"Operation"`
	Data       string `json:"Data"`
}
