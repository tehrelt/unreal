package entity

type From struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Message struct {
	Id       uint32 `json:"id"`
	From     From   `json:"from"`
	Subject  string `json:"subject"`
	Body     string `json:"body"`
	SentDate string `json:"sentDate"`
	IsRead   bool   `json:"isRead"`
}
