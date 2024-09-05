package ns

import (
	"fmt"
	"strings"

	"github.com/OLProtocol/ordx/common"

	"github.com/dgraph-io/badger/v4"
)

func loadNameFromDB(name string, value *NameValueInDB, txn *badger.Txn) error {
	key := GetNameKey(name)
	// return common.GetValueFromDB([]byte(key), txn, value)
	return common.GetValueFromDBWithProto3([]byte(key), txn, value)
}

func GetNameKey(name string) string {
	return fmt.Sprintf("%s%s", DB_PREFIX_NAME, strings.ToLower(name))
}

func GetKVKey(name, key string) string {
	return fmt.Sprintf("%s%s-%s", DB_PREFIX_KV, strings.ToLower(name), key)
}

func ParseNameKey(input string) (string, error) {
	if !strings.HasPrefix(input, DB_PREFIX_NAME) {
		return "", fmt.Errorf("invalid string format")
	}
	str := strings.TrimPrefix(input, DB_PREFIX_NAME)
	return str, nil
}

func ParseKVKey(input string) (string, string, error) {
	if !strings.HasPrefix(input, DB_PREFIX_KV) {
		return "", "", fmt.Errorf("invalid string format")
	}
	str := strings.TrimPrefix(input, DB_PREFIX_KV)
	parts := strings.Split(str, "-")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid string format")
	}

	return parts[0], parts[1], nil
}
