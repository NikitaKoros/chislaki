package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	password := "password654321"
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	fmt.Println(hash)
}
