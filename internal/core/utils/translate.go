package utils

import (
	gt "github.com/bas24/googletranslatefree"
)

// Translate with google translate
func Translate(text, sourceLang, targetLang string) string {
	result, _ := gt.Translate(text, sourceLang, targetLang)
	return result
}
