package utils

import (
	"encoding/base64"
	"net/url"
)

// DecodeStringBase64 decode string base64
func DecodeStringBase64(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return string(data), err
	}
	return string(data), nil
}

// DecodeParam decode param
func DecodeParam(s string) string {
	data, _ := url.QueryUnescape(s)
	return data
}
