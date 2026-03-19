package service

import (
	"crypto/sha256"
	"math/big"
	"strconv"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/domain"
)

func GenerateShortCode(normalizedOriginalURL string, nonce int) string { // Генератор 10значного кода с помощью sha256 и base63
	input := normalizedOriginalURL + "|" + strconv.Itoa(nonce)
	sum := sha256.Sum256([]byte(input))
	value := new(big.Int).SetBytes(sum[:])

	alphabet := domain.ShortCodeAlphabet()
	base := big.NewInt(int64(len(alphabet)))
	result := make([]byte, domain.ShortCodeLength)
	mod := new(big.Int)

	for i := domain.ShortCodeLength - 1; i >= 0; i-- { // Слева направо или наоборот роли не играет, но в данном случае принято писать справа налево
		value.DivMod(value, base, mod) // Аналог % для bigint. Находим с помощью хэша значение в base63
		result[i] = alphabet[mod.Int64()]
	}

	return string(result)
}
