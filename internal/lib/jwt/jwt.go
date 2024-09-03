package jwt

type JWT struct {
	Private []byte
	Public  []byte
}

func NewJWT(privateKey, publicKey []byte) *JWT {
	return &JWT{
		Private: privateKey,
		Public:  publicKey,
	}
}
