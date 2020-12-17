package mainther

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"os"
)

-en

import (
	"crypto/rsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
)

func encryptOAEP() {
	secretMessage := []byte("send reinforcements, we're going to advance")
	label := []byte("orders")

	// crypto/rand.Reader is a good source of entropy for randomizing the
	// encryption function.
	rng := rand.Reader

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, &test2048Key.PublicKey, secretMessage, label)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from encryption: %s\n", err)
		return
	}

	// Since encryption is a randomized function, ciphertext will be
	// different each time.
	fmt.Printf("Ciphertext: %x\n", ciphertext)
	text, _ := rsa.DecryptOAEP(sha256.New(), rng, test2048Key, ciphertext, label)
	fmt.Printf("Plain text: %s\n", text)
}

var test2048Key *rsa.PrivateKey

func main() {
	var err error
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	test2048Key = key

	encryptOAEP()
}
