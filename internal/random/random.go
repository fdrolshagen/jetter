package random

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func Execute(funcName string, arg string) (string, error) {
	switch funcName {
	case "hexadecimal":
		return hexadecimal(arg)
	case "uuid":
		return uuid()
	default:
		return "", fmt.Errorf("unsupported random function: %s", funcName)
	}
}

func hexadecimal(arg string) (string, error) {
	length, err := strconv.Atoi(arg)
	if err != nil {
		return "", errors.New("invalid argument for random.hexadecimal, must be integer")
	}
	if length <= 0 {
		return "", errors.New("length must be > 0")
	}

	bytes := make([]byte, (length+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	hexStr := hex.EncodeToString(bytes)
	hexStr = strings.ToUpper(hexStr)
	return hexStr[:length], nil
}

func uuid() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant is 10

	uuidStr := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)

	return uuidStr, nil
}
