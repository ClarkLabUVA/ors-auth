package main

import (
  jose "github.com/square/go-jose"
  "encoding/json"
  "crypto/rsa"
  "time"
)



func (u User)NewToken(issuer, audience string) (uc UserClaims) {


  uc = UserClaims{
    Id: u.Id,
    IssuedAt: time.Now().Unix(),
    ExpiresAt: time.Now().Add(time.Minute * 10).Unix(),
    Issuer: issuer,
    Subject: u.Id,
    Audience: audience,
    Name: u.Name,
    Email: u.Email,
    Admin: u.Admin,
    Groups: u.Groups,
  }


  // TODO: update user mongo record with

  return
}


type UserClaims struct {
  Id string  `json:"id"`
  IssuedAt int64 `json:"iat"`
  ExpiresAt int64 `json:"exp"`
  Issuer string   `json:"iss"`
  Subject string   `json:"sub"`
  Audience string `json:"aud"`
  Name string `json:"name"`
  Email string `json:"email"`
  Admin bool `json:"is_admin"`
  Groups []string `json:"groups"`
}


// Serialize and Encrypt JOSE
func EncryptJOSE(privateKey *rsa.PrivateKey, uc UserClaims) (serializedToken []byte, err error) {

  token, err := json.Marshal(uc)

  if err != nil {
    return
  }

  publicKey := &privateKey.PublicKey

  encrypter, err := jose.NewEncrypter(
    jose.A128GCM,
    jose.Recipient{Algorithm: jose.RSA_OAEP, Key: publicKey},
    nil,
  )

  if err != nil {
    return
  }

  encryptedToken, err := encrypter.Encrypt(token)
  if err != nil {
    return
  }

  serializedToken = []byte(encryptedToken.FullSerialize())

  return
}


// Parse and Decrypt JOSE
func DecryptJOSE(privateKey *rsa.PrivateKey, joseToken []byte) (uc UserClaims, err error) {

  parsedToken, err := jose.ParseEncrypted(string(joseToken))
  if err != nil {
    return
  }

  decryptedToken, err := parsedToken.Decrypt(privateKey)
  if err != nil {
    return
  }

  err = json.Unmarshal(decryptedToken, &uc)
  if err != nil {
    return
  }

  return

}
