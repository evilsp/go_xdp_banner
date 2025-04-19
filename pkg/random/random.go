package random

import (
	"math/rand"
	"time"
)

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func RandomString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // 创建一个本地随机数生成器
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))] // 从 letters 中随机选择一个字符
	}
	return string(b)
}
