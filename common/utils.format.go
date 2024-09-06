// ConvertTimestampToISO8601 将时间戳转换为 ISO 8601 格式的字符串
package common

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseUtxo(utxo string) (txid string, vout int, err error) {
	parts := strings.Split(utxo, ":")
	if len(parts) != 2 {
		return txid, vout, fmt.Errorf("invalid utxo")
	}

	txid = parts[0]
	vout, err = strconv.Atoi(parts[1])
	if err != nil {
		return txid, vout, err
	}
	if vout < 0 {
		return txid, vout, fmt.Errorf("invalid vout")
	}
	return txid, vout, err
}
