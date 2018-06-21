package main

import (
  "crypto/ecdsa"
  "bytes"
  "crypto/sha256"
  "crypto/elliptic"
  "crypto/rand"
  "log"
  "golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {
  PrivateKey ecdsa.PrivateKey
  PublicKey []byte
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
  private, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
  if err != nil {
    log.Panic(err)
  }
  return *private, append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
}

func NewWallet() *Wallet {
	private, public := newKeyPair()
	return &Wallet{private, public}
}

func HashPubKey(pubKey []byte) []byte {
  publicSHA256 := sha256.Sum256(pubKey)

  RIPEMD160Hasher := ripemd160.New()
  _, err := RIPEMD160Hasher.Write(publicSHA256[:])
  if err != nil {
    log.Panic(err)
  }
  return RIPEMD160Hasher.Sum(nil)
}

func (w Wallet) GetAddress() []byte {
  versionedPayload := append([]byte{version}, HashPubKey(w.PublicKey)...)
  fullPayload := append(versionedPayload, checksum(versionedPayload)...)
  return Base58Encode(fullPayload)
}

func isAddressValid(address string) bool {
  pubKeyHash := Base58Decode([]byte(address))
  actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
  version := pubKeyHash[0]
  pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
  targetChecksum := checksum(append([]byte{version}, pubKeyHash...))
  return bytes.Compare(actualChecksum, targetChecksum) == 0
}

func checksum(payload []byte) []byte {
  firstSHA := sha256.Sum256(payload)
  secondSHA := sha256.Sum256(firstSHA[:])
  return secondSHA[:addressChecksumLen]
}
