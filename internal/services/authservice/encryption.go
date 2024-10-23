package authservice

func (as *Service) encrypt(in string) (string, error) {
	out, err := as.encryptor.Encrypt(in)
	if err != nil {
		return "", err
	}

	return out, nil
}
func (as *Service) decrypt(in string) (string, error) {
	out, err := as.encryptor.Decrypt(in)
	if err != nil {
		return "", err
	}

	return out, nil
}
