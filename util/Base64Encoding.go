package util

import (
	"encoding/base64"
)

// converts base64 back to binary
func DecodeBase64ToBytes(base64Data string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func EncodeBytesToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
