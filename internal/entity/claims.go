package entity

type Claims struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}
