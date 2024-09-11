package entity

type Claims struct {
	Email    string     `json:"email"`
	Password string     `json:"password"`
	Imap     Connection `json:"imap"`
	Smtp     Connection `json:"smtp"`
}
