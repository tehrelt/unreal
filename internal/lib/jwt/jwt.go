package jwt

type JWT struct {
	privateKey []byte
	publicKey  []byte
}

func NewJWT(privateKey, publicKey []byte) *JWT {
	return &JWT{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}
