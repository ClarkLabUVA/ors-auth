package main

import (
  "testing"
  "encoding/json"
  "crypto/rand"
  "crypto/rsa"
  jose "github.com/square/go-jose"
)

func TestFullJWE(t *testing.T) {

  uc := User{
    Id: "orcid:1234-1234-1234-1234",
    Name: "Joe Schmoe",
    Email: "joe.schmoe@example.org",
    Admin: false,
    Groups: []string{},
  }.NewToken("test", "test")

  token, err := json.Marshal(uc)

  if err != nil {
    t.Fatalf("Failed To Serialize UserClaims: %s", err.Error())
  }

  privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
  if err != nil {
    t.Fatalf("Failed to Generate RSA Keypair: %s", err.Error())
  }

  publicKey := &privateKey.PublicKey

  encrypter, err := jose.NewEncrypter(
    jose.A128GCM,
    jose.Recipient{Algorithm: jose.RSA_OAEP, Key: publicKey},
    nil,
  )

  if err != nil {
    t.Fatalf("Failed to Create JOSE Encrypter: %s", err.Error())
  }

  encrypted_token, err := encrypter.Encrypt(token)
  if err != nil {
    t.Fatalf("Failed to Encrypt the Token: %s", err.Error())
  }

  serialized := encrypted_token.FullSerialize()
  t.Logf("Fully Serialized JOSE Token: %s", serialized)

  // attempt to decrypt
  parsed_token, err := jose.ParseEncrypted(serialized)
  if err != nil {
    t.Fatalf("Failed to Parse Serialized JOSE Token: %s", err.Error())
  }

  decrypted_token, err := parsed_token.Decrypt(privateKey)
  if err != nil {
    t.Fatalf("Failed to Decrypt Token: %s", err.Error())
  }

  var decrypted_claims UserClaims
  err = json.Unmarshal(decrypted_token, &decrypted_claims)
  if err != nil {
    t.Fatalf("Failed to Unmarshal Decrypted Token: %s", err.Error())
  }

  t.Logf("Successfully Decrypted Token: %+v", decrypted_claims)

}


func TestJOSE(t *testing.T) {

  // generate a new RSA key for testing
  privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
  if err != nil {
    t.Fatalf("Failed to Generate RSA Keypair: %s", err.Error())
  }

  var serializedToken []byte
  var uc UserClaims
  var u = User{
    Id: "orcid:1234-1234-1234-1234",
    Name: "Joe Schmoe",
    Email: "JoeSchmoe@example.org",
    Admin: false,
    Groups: []string{},
  }

  t.Run("NewToken", func(t *testing.T) {
    uc = u.NewToken("test", "test")

    if uc.Id != u.Id {
      t.Fatalf("Claims Mismatch User Fields\nUser: %+v \nUserClaims %+v", u, uc)
    }

  })

  t.Run("EncryptJOSE", func(t *testing.T){
    serializedToken, err = EncryptJOSE(privateKey, uc)

    if err != nil {
      t.Fatalf("Failed to Encrypt JOSE token: %s", err.Error())
    }

    t.Logf("Serialized Token: %s", string(serializedToken))

  })

  t.Run("DecryptJOSE", func(t *testing.T){
    decryptedClaims, err := DecryptJOSE(privateKey, serializedToken)

    if err != nil {
      t.Fatalf("Failed to Decrypt JOSE Token: %s", err.Error())
    }

    t.Logf("Decrypted Claims: %+v", decryptedClaims)

  })

}
