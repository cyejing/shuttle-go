package utils

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// SHA224String sha
func SHA224String(password string) string {
	hash := sha256.New224()
	hash.Write([]byte(password))
	val := hash.Sum(nil)
	str := ""
	for _, v := range val {
		str += fmt.Sprintf("%02x", v)
	}
	return str
}

func SplitPathAndFile(s string) (string, string, error) {
	if s == "" {
		return "", "", fmt.Errorf("path can not empty")
	}
	slash := strings.LastIndex(s, "/") // -1 if "/" not found
	if slash == -1 {
		return "", s, nil
	}
	if slash == len(s)-1 {
		return s, "", fmt.Errorf("path is illegal : %s", s)
	}
	return s[:slash], s[slash+1:], nil
}
