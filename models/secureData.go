package models

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// SecureData is used to store and send access controlled Param values
// to the locations they are needed at.  SecureData uses a simple
// encryption mechanism based on the NACL Box API (as implemented by
// libsodium, golang.org/x/crypto/nacl/box, and many others).
//
// All fields in this struct will be marshalled into JSON as base64
// encoded strings.
//
// swagger:model
type SecureData struct {
	// Key is the ephemeral public key created by Seal().  It must not
	// be modified after Seal() has completed, and it must be 32 bytes
	// long.
	Key []byte
	// Nonce must be 24 bytes of cryptographically random numbers.  It is
	// populated by Seal(), and must not be modified afterwards.
	Nonce []byte
	// Payload is the encrypted payload generated by Seal().  It must
	// not be modified, and will be 16 bytes longer than the unencrypted
	// data.
	Payload []byte
}

var (
	BadKey   = errors.New("Key must be 32 bytes long")
	BadNonce = errors.New("Nonce must be 24 bytes long")
	Corrupt  = errors.New("SecureData corrupted")
)

// Validate makes sure that the lengths we expect for the Key and
// Nonce are correct.
func (s *SecureData) Validate() error {
	if len(s.Key) != 32 {
		return BadKey
	}
	if len(s.Nonce) != 24 {
		return BadNonce
	}
	if len(s.Payload) < box.Overhead {
		return Corrupt
	}
	return nil
}

// Seal takes curve25519 public key advertised by where the payload
// should be stored, and fills in the SecureData with the data
// required for the Open operation to succeed.
//
// Seal performs the following operations internally:
//
// * Generate ephemeral cuve25519 public and private keys from the
//   system's cryptographically secure random number generator.
//
// * Generate a 24 byte nonce from the same random number generator used
//   to create the keys.
//
// * Encrypt the data using the peer public key, the ephemeral private key,
//   and the generated nonce.
//
// * Populate s.Key with the ephemeral public key, s.Nonce with the
//   generated nonce, and s.Payload with the encrypted data.
func (s *SecureData) Seal(peerPublicKey *[32]byte, data []byte) error {
	ourPublicKey, ourPrivateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("Error generating ephemeral local keys: %v", err)
	}
	s.Key = ourPublicKey[:]
	nonce := [24]byte{}
	_, err = io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return fmt.Errorf("Error generating nonce: %v", err)
	}
	s.Nonce = nonce[:]
	s.Payload = box.Seal(nil, data, &nonce, peerPublicKey, ourPrivateKey)
	return nil
}

// Marshal marshals the passed-in data into JSON and calls Seal() with
// peerPublicKey and the marshalled data.
func (s *SecureData) Marshal(peerPublicKey []byte, data interface{}) error {
	if len(peerPublicKey) != 32 {
		return BadKey
	}
	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}
	ppk := [32]byte{}
	copy(ppk[:], peerPublicKey)
	return s.Seal(&ppk, buf)
}

// Open opens a sealed SecureData item.  It takes the private key that
// matches the peerPublicKey passed to the Seal() operation.  If the
// SecureData object has not been corrupted or tampered with, Open()
// will return the decrypted data , otherwise it will return an error.
//
// Open performs the following operations internally:
//
// * Validate that the lengths of all the fields of the SecureData
//   struct are within expected ranges.  This is not required for
//   correctness, but it alows us to return nicer errors.
//
// * Extract the stored ephemeral public key and nonce from the
//   SecureData object.
//
// * Decrypt Payload using the extracted nonce, the extracted public key,
//   and the passed-in private key.
//
// * If any errors were returned in the decrypt process, return a
//   Corrupt error, otherwise return the decrypted data.
func (s *SecureData) Open(targetPrivateKey *[32]byte) ([]byte, error) {
	err := s.Validate()
	if err != nil {
		return nil, err
	}
	peerPublicKey := [32]byte{}
	copy(peerPublicKey[:], s.Key[:])
	nonce := [24]byte{}
	copy(nonce[:], s.Nonce[:])
	res, opened := box.Open(nil, s.Payload, &nonce, &peerPublicKey, targetPrivateKey)
	if !opened {
		return nil, Corrupt
	}
	return res, nil
}

func (s *SecureData) Unmarshal(targetPrivateKey []byte, res interface{}) error {
	if len(targetPrivateKey) != 32 {
		return BadKey
	}
	tpk := [32]byte{}
	copy(tpk[:], targetPrivateKey)
	buf, err := s.Open(&tpk)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, res)
}
