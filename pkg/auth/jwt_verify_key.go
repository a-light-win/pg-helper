package auth

import (
	"fmt"
	"sync"

	"github.com/a-light-win/pg-helper/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

type JwtVerifyKey struct {
	Type []string `enum:"ES256,ES384,ES512,EdDSA" help:"The type of the key to use for verification"`
	File []string `validate:"samelen=Type" type:"file" help:"The path to the key file to use for verification"`

	keyLock sync.Mutex
	keys    map[string]interface{}
}

func (k *JwtVerifyKey) LoadVerifyKey(token *jwt.Token) (interface{}, error) {
	if key, ok := k.keys[token.Method.Alg()]; ok {
		return key, nil
	}

	k.keyLock.Lock()
	defer k.keyLock.Unlock()

	for i := range k.Type {
		if k.Type[i] == token.Method.Alg() {
			return k.loadKey(i)
		}
	}
	return nil, fmt.Errorf("no key found for algorithm %s", token.Method.Alg())
}

func (k *JwtVerifyKey) loadKey(index int) (interface{}, error) {
	type_ := k.Type[index]
	if key, ok := k.keys[type_]; ok {
		return key, nil
	}

	key, err := utils.LoadPublicKey(k.File[index])
	if err != nil {
		return nil, err
	}
	k.keys[type_] = key
	return key, nil
}

func (k *JwtVerifyKey) ResetKey() {
	k.keyLock.Lock()
	defer k.keyLock.Unlock()

	k.keys = make(map[string]interface{})
}
