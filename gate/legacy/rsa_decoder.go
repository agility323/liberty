package legacy

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

type RsaDecoder struct {
	privateKey *rsa.PrivateKey
}

var rsaDecoder = &RsaDecoder{}

func InitRsaKey(path string) error{
	return rsaDecoder.parsePrivateKeyFromFile(path)
}


func (decoder *RsaDecoder) parsePrivateKeyFromFile(path string) error{
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.New(fmt.Sprintf("RSA Private Key Invalid: %s", err.Error()))
	}
	if dataBlock, _ := pem.Decode(fileData); dataBlock == nil {
		return errors.New(fmt.Sprintf("RSA Private Key Invalid: pem decode failed"))
	} else if key, err := x509.ParsePKCS1PrivateKey(dataBlock.Bytes); err != nil {
		return errors.New(fmt.Sprintf("RSA Private Key Invalid: %s", err.Error()))
	} else {
		decoder.privateKey = key
	}
	return nil
}

func (decoder *RsaDecoder) decode(data []byte) ([]byte, error) {
	return rsa.DecryptOAEP(crypto.SHA1.New(), rand.Reader, decoder.privateKey, data, nil)
}
