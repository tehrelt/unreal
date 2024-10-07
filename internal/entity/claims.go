package entity

type Claims struct {
	Id string `json:"id"`
}

type SessionInfo struct {
	Email    string     `json:"email"`
	Password string     `json:"password"`
	Imap     Connection `json:"imap"`
	Smtp     Connection `json:"smtp"`
}
