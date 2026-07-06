package activity

import (
	"sync"
)

const (
	Arguments = "arguments"
	Responses = "responses"
)

var (
	setNameLocker     sync.Mutex //避免冲突
	returnKeyPrefix   = ""
	keyPrefixLinkChar = "/"
)

func ReturnKeyPrefix() string {
	return returnKeyPrefix
}
func WithReturnKeyPrefix(keyPrefix string) {
	if returnKeyPrefix == keyPrefix {
		return
	}
	setNameLocker.Lock()
	defer setNameLocker.Unlock()
	returnKeyPrefix = keyPrefix
}
