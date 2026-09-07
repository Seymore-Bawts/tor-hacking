#!/bin/bash

set -e

KEY_DIR="./keys"
PUB_KEY="$KEY_DIR/public_key.pem"
PRIV_KEY="$KEY_DIR/private_key.pem"

echo "=== Generating RSA Keys ==="

mkdir -p "$KEY_DIR"

# Generate RSA key pair
if command -v openssl &> /dev/null; then
  openssl genrsa -out "$PRIV_KEY" 2048 2>/dev/null
  openssl rsa -in "$PRIV_KEY" -pubout -out "$PUB_KEY" 2>/dev/null
  echo "Keys generated using openssl"
else
  # Fallback: generate via Go
  cat > "$KEY_DIR/gen_keys.go" << 'EOF'
package main

import (
  "crypto/rand"
  "crypto/rsa"
  "crypto/x509"
  "encoding/pem"
  "os"
)

func main() {
  priv, err := rsa.GenerateKey(rand.Reader, 2048)
  if err != nil {
    panic(err)
  }

  privBytes, err := x509.MarshalPKCS1PrivateKey(priv)
  if err != nil {
    panic(err)
  }
  
  privPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
  os.WriteFile("keys/private_key.pem", privPem, 0600)

  pubBytes, err := x509.MarshalPKCS1PublicKey(&priv.PublicKey)
  if err != nil {
    panic(err)
  }
  
  pubPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubBytes})
  os.WriteFile("keys/public_key.pem", pubPem, 0644)
}
EOF
  cd "$KEY_DIR" && go run gen_keys.go && cd - > /dev/null
  echo "Keys generated using Go"
fi

echo "Public key: $PUB_KEY"
echo "Private key: $PRIV_KEY"