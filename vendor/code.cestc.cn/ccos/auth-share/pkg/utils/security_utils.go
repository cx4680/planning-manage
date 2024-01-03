package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"hash"
	"io"
	"strings"
)

const (
	pemPriBegin = "-----BEGIN RSA PRIVATE KEY-----"
	pemPriEnd   = "-----END RSA PRIVATE KEY-----"
	pemPubBegin = "-----BEGIN PUBLIC KEY-----"
	pemPubEnd   = "-----END PUBLIC KEY-----"
)

func AesSecret() (string, error) {
	key := make([]byte, 16)

	_, err := rand.Read(key)

	if err != nil {
		return "", errors.New("create secret error")
	}

	secret := base64.StdEncoding.EncodeToString(key)

	return secret, nil
}

func AesDecrypt(base64Content string, secret string) (string, error) {

	content, err := base64.StdEncoding.DecodeString(base64Content)

	if len(content) < 12+16 {
		return "", errors.New("非法参数异常")
	}

	block, err := aes.NewCipher([]byte(secret))

	if err != nil {
		return "", err
	}

	nonce, cipherByte := content[:12], content[12:]

	aesGcm, err := cipher.NewGCM(block)

	if err != nil {
		return "", err
	}

	plainByte, err := aesGcm.Open(nil, nonce, cipherByte, nil)
	if err != nil {
		return "", err
	}

	return string(plainByte), nil
}

func RSADecryptBySecretStr(content []byte, privateKeyStr string) ([]byte, error) {

	privateKey, err := ParsePrivateKey(privateKeyStr)

	decryptContent, err := RSADecrypt(content, privateKey)

	return decryptContent, err
}

func RSAEncryptBySecretStr(content []byte, publicKeyStr string) ([]byte, error) {

	publicKey, err := ParsePublicKey(publicKeyStr)

	encryptContent, err := RSAEncrypt(content, publicKey)

	return encryptContent, err
}

func RSADecrypt(content []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	result, err := DecryptOAEP(sha1.New(), rand.Reader, privateKey, content, nil)

	return result, err
}

func DecryptOAEP(hash hash.Hash, random io.Reader, private *rsa.PrivateKey, msg []byte, label []byte) ([]byte, error) {
	msgLen := len(msg)
	step := private.PublicKey.Size()
	var decryptedBytes []byte
	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}
		decryptedBlockBytes, err := rsa.DecryptOAEP(hash, random, private, msg[start:finish], label)
		if err != nil {
			return nil, err
		}

		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}
	return decryptedBytes, nil
}

func RSAEncrypt(content []byte, publicKey *rsa.PublicKey) ([]byte, error) {

	encryptedBytes, err := EncryptOAEP(sha1.New(), rand.Reader, publicKey, content, nil)

	return encryptedBytes, err
}

func EncryptOAEP(hash hash.Hash, random io.Reader, public *rsa.PublicKey, msg []byte, label []byte) ([]byte, error) {
	msgLen := len(msg)
	step := public.Size() - 2*hash.Size() - 2
	var encryptedBytes []byte
	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}
		encryptedBlockBytes, err := rsa.EncryptOAEP(hash, random, public, msg[start:finish], label)
		if err != nil {
			return nil, err
		}
		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)
	}

	return encryptedBytes, nil
}

func ParsePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	privateKey = FormatKeyStr(privateKey, true)

	block, _ := pem.Decode([]byte(privateKey))

	if block == nil {
		return nil, errors.New("私钥信息错误！")
	}

	priKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	return priKey, nil
}

func ParsePublicKey(publicKey string) (*rsa.PublicKey, error) {

	publicKey = FormatKeyStr(publicKey, false)

	block, _ := pem.Decode([]byte(publicKey))

	if block == nil {
		return nil, errors.New("公钥信息错误！")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	return publicKeyInterface.(*rsa.PublicKey), nil
}

func FormatKeyStr(keyStr string, isPri bool) string {
	if isPri {
		if !strings.Contains(keyStr, pemPriBegin) {
			keyStr = pemPriBegin + "\n" + keyStr
		}
		if !strings.Contains(keyStr, pemPriEnd) {
			keyStr = keyStr + "\n" + pemPriEnd
		}
	}

	if !isPri {
		if !strings.Contains(keyStr, pemPubBegin) {
			keyStr = pemPubBegin + "\n" + keyStr
		}
		if !strings.Contains(keyStr, pemPubEnd) {
			keyStr = keyStr + "\n" + pemPubEnd
		}
	}

	return keyStr
}
