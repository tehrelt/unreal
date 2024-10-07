package authservice

func (as *AuthService) encrypt(in string) (string, error) {
	out, err := as.encryptor.Encrypt(in)
	if err != nil {
		return "", err
	}

	return out, nil
}
func (as *AuthService) decrypt(in string) (string, error) {
	out, err := as.encryptor.Decrypt(in)
	if err != nil {
		return "", err
	}

	return out, nil
}
