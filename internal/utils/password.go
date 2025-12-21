package utils

import gonanoid "github.com/matoous/go-nanoid/v2"

const PASSWORD_CHARSET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateNewPassword(size int) string {
	return gonanoid.MustGenerate(PASSWORD_CHARSET, size)
}
