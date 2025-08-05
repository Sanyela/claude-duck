package utils

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// GenerateSecurePassword 生成安全的随机密码
// 长度: 15-20位 (随机)
// 包含: 大写字母、小写字母、数字、特殊字符
func GenerateSecurePassword() (string, error) {
	// 随机长度 15-20 位
	lengthRange := 6 // 20 - 15 + 1
	lengthOffset, err := rand.Int(rand.Reader, big.NewInt(int64(lengthRange)))
	if err != nil {
		return "", err
	}
	length := 15 + int(lengthOffset.Int64())

	// 字符集定义
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits := "0123456789"
	special := "!@#$%^&*()-_=+[]{}|;:,.<>?"

	// 确保密码包含每种字符类型至少一个
	var password strings.Builder
	charSets := []string{lowercase, uppercase, digits, special}
	
	// 每种类型至少一个字符
	for _, charset := range charSets {
		char, err := getRandomChar(charset)
		if err != nil {
			return "", err
		}
		password.WriteString(char)
	}

	// 剩余位数随机填充
	allChars := lowercase + uppercase + digits + special
	for password.Len() < length {
		char, err := getRandomChar(allChars)
		if err != nil {
			return "", err
		}
		password.WriteString(char)
	}

	// 随机打乱密码字符顺序
	passwordStr := password.String()
	return shuffleString(passwordStr)
}

// getRandomChar 从字符集中获取随机字符
func getRandomChar(charset string) (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return "", err
	}
	return string(charset[n.Int64()]), nil
}

// shuffleString 随机打乱字符串
func shuffleString(s string) (string, error) {
	chars := []rune(s)
	for i := len(chars) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		chars[i], chars[j.Int64()] = chars[j.Int64()], chars[i]
	}
	return string(chars), nil
}