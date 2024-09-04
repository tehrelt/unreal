package entity

type Message struct {
	Id      string `json:"id"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
