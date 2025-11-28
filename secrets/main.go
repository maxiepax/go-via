package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Credits to this person for excellent code: https://www.melvinvivas.com/how-to-encrypt-and-decrypt-data-using-aes/

func Init() string {
	var key []byte
	if _, err := os.Stat("secret/secret.key"); os.IsNotExist(err) {
		//secrets file does not exist, create folder and file
		err = os.MkdirAll("secret", os.ModePerm)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Warn("could not create secrets directory")
		}
		logrus.WithFields(logrus.Fields{
			"key": "no secrets file has been detected, attempting to create a new one and generate secret key",
		}).Info("secrets")
		file, err := os.Create("secret/secret.key")
		if err != nil {
			logrus.Fatal(err.Error())
		}
		logrus.WithFields(logrus.Fields{
			"key": "secret/secret.key created",
		}).Info("secrets")

		//generate a random 32 byte AES-256 key
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			panic(err.Error())
		}

		//convert key to string and write to file
		hexkey := hex.EncodeToString(bytes)
		wr, err := file.WriteString(hexkey)
		if err != nil {
			logrus.Fatal(err.Error())
		}
		logrus.WithFields(logrus.Fields{
			"key":   "secret key persisted to file",
			"bytes": wr,
		}).Info("secrets")
		err = file.Close()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Warn("could not close secrets file")
		}
		key, _ = os.ReadFile("secret/secret.key")
	} else {
		//Database exists, moving on.
		logrus.WithFields(logrus.Fields{
			"key": "found existing secret key!",
		}).Info("secrets")
		key, _ = os.ReadFile("secret/secret.key")
	}
	return string(key)
}

func Encrypt(stringToEncrypt string, keyString string) (encryptedString string) {

	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(keyString)
	plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext)
}

func Decrypt(encryptedString string, keyString string) (decryptedString string) {

	key, _ := hex.DecodeString(keyString)
	enc, _ := hex.DecodeString(encryptedString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return string(plaintext)
}
