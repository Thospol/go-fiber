package utils

import (
	"encoding/base64"
	"net/url"
)

// EncodeStringBase64 encode string base64
func EncodeStringBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// EncodeParam endcode param
func EncodeParam(s string) string {
	return url.QueryEscape(s)
}
