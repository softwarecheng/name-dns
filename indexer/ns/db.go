package ns

import (
	"fmt"
	"strings"

	"github.com/OLProtocol/ordx/common"

	"github.com/dgraph-io/badger/v4"
)

func loadNameFromDB(name string, value *NameValueInDB, txn *badger.Txn) error {
	key := GetNameKey(name)
	return common.GetValueFromDBWithProto3([]byte(key), txn, value)
}

func GetNameKey(name string) string {
	return fmt.Sprintf("%s%s", DB_PREFIX_NAME, strings.ToLower(name))
}
