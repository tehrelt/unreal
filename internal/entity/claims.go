package entity

type Claims struct {
	EMail          string `json:"email"`
	HashedPassword string `json:"hashedPassword"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
}
