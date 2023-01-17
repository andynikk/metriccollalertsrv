package encryption

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
)

type KeyEncryption struct {
	TypeEncryption string
	PublicKey      *rsa.PublicKey
	PrivateKey     *rsa.PrivateKey
}

func (key *KeyEncryption) RsaEncrypt(msg []byte) ([]byte, error) {
	if key == nil {
		return msg, nil
	}
	encryptedBytes, err := rsa.EncryptOAEP(sha512.New512_256(), rand.Reader, key.PublicKey, msg, nil)
	return encryptedBytes, err
}

func (key *KeyEncryption) RsaDecrypt(msgByte []byte) ([]byte, error) {
	if key == nil {
		return msgByte, nil
	}
	msgByte, err := key.PrivateKey.Decrypt(nil, msgByte, &rsa.OAEPOptions{Hash: crypto.SHA512_256})
	return msgByte, err
}

func CreateCert() ([]bytes.Buffer, error) {
	var numSert int64
	var subjectKeyID string
	var lenKeyByte int

	fmt.Print("Введите уникальный номер сертификата: ")
	if _, err := fmt.Fscan(os.Stdin, &numSert); err != nil {
		constants.Logger.ErrorLog(err)
		return nil, err
	}

	fmt.Print("Введите ИД ключа субъекта (пример ввода 12346): ")
	if _, err := fmt.Fscan(os.Stdin, &subjectKeyID); err != nil {
		constants.Logger.ErrorLog(err)
		return nil, err
	}

	fmt.Print("Длина ключа в байтах: ")
	if _, err := fmt.Fscan(os.Stdin, &lenKeyByte); err != nil {
		constants.Logger.ErrorLog(err)
		return nil, err
	}

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(numSert),
		Subject: pkix.Name{
			Organization: []string{"AdvancedMetrics"},
			Country:      []string{"RU"},
		},
		NotBefore: time.Now(),
		NotAfter: time.Now().AddDate(constants.TimeLivingCertificateYaer, constants.TimeLivingCertificateMounth,
			constants.TimeLivingCertificateDay),
		SubjectKeyId: []byte(subjectKeyID),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, lenKeyByte)
	if err != nil {
		return nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	var certPEM bytes.Buffer
	_ = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	var privateKeyPEM bytes.Buffer
	_ = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return []bytes.Buffer{certPEM, privateKeyPEM}, nil
}

func SaveKeyInFile(key *bytes.Buffer, pathFile string) {
	file, err := os.Create(pathFile)
	if err != nil {
		return
	}
	_, err = file.WriteString(key.String())
	if err != nil {
		return
	}
}

func InitPrivateKey(cryptoKeyPath string) (*KeyEncryption, error) {

	if cryptoKeyPath == "" {
		return nil, errors.New("путь к приватному ключу не указан")
	}
	pvkData, _ := os.ReadFile(cryptoKeyPath)
	pvkBlock, _ := pem.Decode(pvkData)
	pvk, err := x509.ParsePKCS1PrivateKey(pvkBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return &KeyEncryption{TypeEncryption: constants.TypeEncryption, PrivateKey: pvk, PublicKey: &pvk.PublicKey}, nil
}

func InitPublicKey(cryptoKeyPath string) (*KeyEncryption, error) {
	if cryptoKeyPath == "" {
		return nil, errors.New("не указан путь к публичному ключу")
	}
	certData, _ := os.ReadFile(cryptoKeyPath)
	certBlock, _ := pem.Decode(certData)
	cert, _ := x509.ParseCertificate(certBlock.Bytes)
	certPublicKey := cert.PublicKey.(*rsa.PublicKey)
	return &KeyEncryption{TypeEncryption: constants.TypeEncryption, PublicKey: certPublicKey}, nil
}
