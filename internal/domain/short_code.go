package domain

import "unicode/utf8"

const (
	ShortCodeLength   = 10
	shortCodeAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

var shortCodeAlphabetSet = buildAlphabetSet(shortCodeAlphabet)

func ValidateShortCode(code string) error { // Проверка длина кода ровно 10 и используются только допустимые символы
	if utf8.RuneCountInString(code) != ShortCodeLength {
		return ErrInvalidShortCode
	}

	for _, r := range code {
		if !shortCodeAlphabetSet[r] {
			return ErrInvalidShortCode
		}
	}

	return nil
}

func ShortCodeAlphabet() string {
	return shortCodeAlphabet
}

func buildAlphabetSet(alphabet string) map[rune]bool { // Сборка допустимых символов
	set := make(map[rune]bool, len(alphabet))
	for _, r := range alphabet {
		set[r] = true
	}

	return set
}
