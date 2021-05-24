package utils

import "regexp"

var (
	regexEmail           = regexp.MustCompile(`^[\w-]+(\.[\w-]+)*@([a-z0-9-]+(\.[a-z0-9-]+)*?\.[a-z]{2,6}|(\d{1,3}\.){3}\d{1,3})(:\d{4})?$`)
	regexIDCard          = regexp.MustCompile(`^(\d{13})?$`)
	regexPhoneNumber     = regexp.MustCompile(`0[6|8|9]{1}\d{8}$`)
	regexTelephoneNumber = regexp.MustCompile(`0[2|3|4|5|7]{1}\d{7}$`)
	regexEnglishAlphabet = regexp.MustCompile(`^[a-zA-Z]+$`)
	regexNumber          = regexp.MustCompile(`^[0-9]+$`)
	regexWhiteSpace      = regexp.MustCompile(`[[:space:]]`)
	regexEngNumber       = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

// IsValidEmail check email is valid
func IsValidEmail(email string) bool {
	return regexEmail.MatchString(email)
}

// IsValidIDCard check ID cadd is valid
func IsValidIDCard(IdCard string) bool {
	return isValidCitizenID(IdCard)
}

// IsValidPhoneNumber check phone number is valid
func IsValidPhoneNumber(phoneNumber string) bool {
	return regexPhoneNumber.MatchString(phoneNumber)
}

// IsValidTelephoneNumber check telephone number is valid
func IsValidTelephoneNumber(telephoneNumber string) bool {
	return regexTelephoneNumber.MatchString(telephoneNumber)
}

// IsValidEnglishAlphabet check english alphabet is valid
func IsValidEnglishAlphabet(text string) bool {
	return regexEnglishAlphabet.MatchString(text)
}

// IsValidNumber check number is valid
func IsValidNumber(number string) bool {
	return regexNumber.MatchString(number)
}

// RemoveWhiteSpaceFromString remove special char from string
func RemoveWhiteSpaceFromString(text string) string {
	return regexWhiteSpace.ReplaceAllString(text, " ")
}

func ReplaceSpecialCharacter(text string) string {
	return regexEngNumber.ReplaceAllString(text, " ")
}
