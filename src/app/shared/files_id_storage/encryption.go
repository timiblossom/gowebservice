package files_id_storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"io/ioutil"
)

// Decrypt do decripting file with aes
func Decrypt(input []byte, keystring string) ([]byte, error) {
	if input == nil {
		return nil, errors.New("error while decrypt info: input is nil")
	}

	if keystring == "" {
		return nil, errors.New("error while decrypt info: keystring is empty")
	}

	// Key
	key := []byte(keystring)

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.New("error while encrypt file: can't crate AES cipher - " + err.Error())
	}

	// Before even testing the decryption,
	// if the text is too small, then it is incorrect
	if len(input) < aes.BlockSize {
		return nil, errors.New("error while encrypt data: data is too short")
	}

	// Get the 16 byte IV
	iv := input[:aes.BlockSize]

	// Remove the IV from the ciphertext
	input = input[aes.BlockSize:]

	// Return a decrypted stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt bytes from ciphertext
	stream.XORKeyStream(input, input)

	return input, nil
}

// Encrypt info with aes
func Encrypt(input []byte, keystring string) ([]byte, error) {
	if input == nil {
		return nil, errors.New("error while encrypt info: input is nil")
	}

	if keystring == "" {
		return nil, errors.New("error while encrypt info: keystring is empty")
	}

	// Key
	key := []byte(keystring)

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.New("error while decrypt file: can't crate AES cipher - " + err.Error())
	}

	// Empty array of 16 + plaintext length
	// Include the IV at the beginning
	output := make([]byte, aes.BlockSize+len(input))

	// Slice of first 16 bytes
	iv := output[:aes.BlockSize]

	// Write 16 rand bytes to fill iv
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, errors.New("errir while decrypt file (read full): " + err.Error())
	}

	// Return an encrypted stream
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt bytes from input to output
	stream.XORKeyStream(output[aes.BlockSize:], input)

	return output, nil
}

// WriteToFile tool func for write data to file
func WriteToFile(data []byte, filePath string) error {
	if data == nil {
		return errors.New("error while write data to file: data is nil")
	}

	if filePath == "" {
		return errors.New("error while write data to file: file path is empty")
	}

	return ioutil.WriteFile(filePath, data, 777)
}

// ReadFromFile tool func for read data from file
func ReadFromFile(filePath string) ([]byte, error) {
	if filePath == "" {
		return nil, errors.New("error while read data from file: file path is empty")
	}

	data, err := ioutil.ReadFile(filePath)
	return data, err
}
