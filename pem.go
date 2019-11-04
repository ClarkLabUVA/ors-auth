package main

import (
  "bufio"
  "encoding/pem"

  "crypto/rsa"
  "crypto/x509"

  //"fmt"
  "os"
)

// TODO
// 1: Error Formatting
func readPem(filepath string) (r *rsa.PrivateKey, err error) {

  keyfile, err := os.Open(filepath)
  defer keyfile.Close()

  if err != nil {
    return
  }

  pemInfo, _ := keyfile.Stat()
  pembytes := make([]byte, pemInfo.Size())

  buffer := bufio.NewReader(keyfile)
  _, err = buffer.Read(pembytes)

  data, _ := pem.Decode([]byte(pembytes))

  parsed, err := x509.ParsePKCS8PrivateKey(data.Bytes)

  r = parsed.(*rsa.PrivateKey)

  return
}
