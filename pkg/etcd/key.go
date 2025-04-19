package etcd

import (
	"path"
	"strings"
)

type Key = string

// Join 拼接键路径，接收多个 Key，返回拼接后的 Key
func Join(elements ...Key) Key {
	return path.Join(elements...)
}

// Split 分割键路径，返回每一级的路径元素作为 Key 切片
func Split(key Key) []Key {
	parts := strings.Split(strings.Trim(key, "/"), "/")
	return parts
}

// Base 返回键路径的最后一级（叶子节点）
func Base(key Key) Key {
	return path.Base(key)
}

// Dir 返回键路径的父级路径
func Dir(key Key) Key {
	return path.Dir(key)
}

// IsAbs 判断键是否是绝对路径
func IsAbs(key Key) bool {
	return strings.HasPrefix(key, "/")
}

// Clean 清理键路径中的多余斜杠，返回一个新的 Key
func Clean(key Key) Key {
	return path.Clean(key)
}
