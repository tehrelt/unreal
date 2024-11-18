package dsa

import (
	dsa "crypto/ed25519"
	"crypto/sha1"
)

type Signer struct {
	public  dsa.PublicKey
	private dsa.PrivateKey
}

func New(public dsa.PublicKey, private dsa.PrivateKey) *Signer {
	return &Signer{
		public:  public,
		private: private,
	}
}

func (s *Signer) Sign(in []byte) []byte {
	hash := sha1.Sum(in)
	signature := dsa.Sign(s.private, hash[:])
	return signature
}

func (s *Signer) Verify(in, signature []byte) error {
	hash := sha1.Sum(in)
	if !dsa.Verify(s.public, hash[:], signature) {
		return ErrSignatureVerificationFailed
	}

	return nil
}
