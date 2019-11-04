package main


import (
  "testing"
)

func TestLoadPem(t *testing.T) {

  _, err := readPem("./ors-auth/server.key")

  if err != nil {
    t.Fatalf("Failed to Read in PEM file: %s", err.Error())
  }

}
