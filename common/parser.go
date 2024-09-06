package common

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/btcsuite/btcd/txscript"
	"github.com/fxamacker/cbor/v2"
)

func GetRawData(txID string, network string) ([][]byte, error) {
	url := ""
	switch network {
	case ChainTestnet:
		url = fmt.Sprintf("https://mempool.space/testnet/api/tx/%s", txID)
	case ChainTestnet4:
		url = fmt.Sprintf("https://mempool.space/testnet4/api/tx/%s", txID)
	case ChainMainnet:
		url = fmt.Sprintf("https://mempool.space/api/tx/%s", txID)
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transaction data for %s from the API, error: %v", txID, err)

	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve transaction data for %s from the API, error: %v", txID, err)
	}

	var data map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON response for %s, error: %v", txID, err)
	}
	txWitness := data["vin"].([]interface{})[0].(map[string]interface{})["witness"].([]interface{})

	if len(txWitness) < 2 {
		return nil, fmt.Errorf("failed to retrieve witness for %s", txID)
	}

	var rawData [][]byte = make([][]byte, len(txWitness))
	for i, v := range txWitness {
		rawData[i], err = hex.DecodeString(v.(string))
		if err != nil {
			return nil, fmt.Errorf("failed to decode hex string to byte array for %s, error: %v", txID, err)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string to byte array for %s, error: %v", txID, err)
	}
	return rawData, nil
}

func readBytes(raw []byte, pointer *int, n int) []byte {
	value := raw[*pointer : *pointer+n]
	*pointer += n
	return value
}

func getBeginPosition(raw []byte) int {
	inscriptionMark := []byte{0, txscript.OP_IF, 3, 111, 114, 100}                         // "0063036f7264"
	inscriptionMark2 := []byte{0, txscript.OP_IF, txscript.OP_PUSHDATA1, 3, 111, 114, 100} // "00634c036f7264"

	position := bytes.Index(raw, inscriptionMark)
	if position >= 0 {
		return int(position + len(inscriptionMark))
	}

	position = bytes.Index(raw, inscriptionMark2)
	if position >= 0 {
		return int(position + len(inscriptionMark2))
	}

	return -1
}

// 从跳过信封头开始
func getEndPosition(raw []byte) int {
	length := len(raw)
	i := 0
	for i < length {
		opcode := raw[i]
		if opcode == txscript.OP_0 {
			i++
			getContentLength(raw, &i)
		} else if txscript.OP_DATA_1 <= opcode && opcode <= txscript.OP_DATA_75 {
			i++
			i += int(opcode)
		} else if opcode >= txscript.OP_PUSHDATA1 && opcode <= txscript.OP_PUSHDATA4 {
			getPushDataLength(raw, &i)
		} else if opcode >= txscript.OP_1 && opcode <= txscript.OP_16 {
			i++
		} else if opcode == txscript.OP_1NEGATE { // testnet: f8fc655ffe139d9952e673c53b7d15cb4b82de5ef036c7fc1211262bbd29bec8
			i++
		} else if opcode == txscript.OP_ENDIF {
			return i
		} else {
			Log.Warnf("unsupport op_code %d ", opcode)
			return -1
		}
	}

	Log.Warnf("not find OP_ENDIF")
	return -2
}

func readPushData(raw []byte, posPointer *int, opcode byte) []byte {
	if txscript.OP_DATA_1 <= opcode && opcode <= txscript.OP_DATA_75 {
		return readBytes(raw, posPointer, int(opcode))
	}

	if opcode >= txscript.OP_1 && opcode <= txscript.OP_16 {
		byt := raw[*posPointer-1] - txscript.OP_1 + 1
		return []byte{byt}
	}

	if opcode == txscript.OP_1NEGATE {
		return []byte{opcode}
	}

	var numBytes int = 0
	var size int = 0
	switch opcode {
	case txscript.OP_PUSHDATA1:
		numBytes = 1
	case txscript.OP_PUSHDATA2:
		numBytes = 2
	case txscript.OP_PUSHDATA4:
		numBytes = 4
	default:
		return nil
	}
	sizeBytes := readBytes(raw, posPointer, numBytes)
	switch opcode {
	case txscript.OP_PUSHDATA1:
		size = int(sizeBytes[0])
	case txscript.OP_PUSHDATA2:
		size = int(binary.LittleEndian.Uint16(sizeBytes))
	case txscript.OP_PUSHDATA4:
		size = int(binary.LittleEndian.Uint32(sizeBytes))
	}
	return readBytes(raw, posPointer, size)
}

func readContent(raw []byte, pos *int) (content []byte, err error) {
	data := []byte{}
	opcode := readBytes(raw, pos, 1)
	if opcode[0] == txscript.OP_ENDIF {
		*pos--
		return nil, nil
	}
	chunk := readPushData(raw, pos, opcode[0])
	for chunk != nil {
		data = append(data, chunk...)
		opcode = readBytes(raw, pos, 1)
		if opcode[0] == txscript.OP_ENDIF {
			*pos--
			break
		} else if opcode[0] == txscript.OP_0 {
			// 某些情况会用OP_0分割，跳过
			opcode = readBytes(raw, pos, 1)
		}
		chunk = readPushData(raw, pos, opcode[0])
		if chunk == nil {
			*pos--
		}
	}
	return data, nil
}

func getPushDataLength(raw []byte, posPointer *int) int {
	opcode := raw[*posPointer]
	if txscript.OP_DATA_1 <= opcode && opcode <= txscript.OP_DATA_75 {
		*posPointer++
		*posPointer += int(opcode)
		return int(opcode)
	}

	if opcode >= txscript.OP_1 && opcode <= txscript.OP_16 {
		*posPointer++
		return 1
	}

	var numBytes int = 0
	var size int = 0
	switch opcode {
	case txscript.OP_PUSHDATA1:
		numBytes = 1
	case txscript.OP_PUSHDATA2:
		numBytes = 2
	case txscript.OP_PUSHDATA4:
		numBytes = 4
	default:
		return 0
	}
	*posPointer++
	sizeBytes := readBytes(raw, posPointer, numBytes)
	switch opcode {
	case txscript.OP_PUSHDATA1:
		size = int(sizeBytes[0])
	case txscript.OP_PUSHDATA2:
		size = int(binary.LittleEndian.Uint16(sizeBytes))
	case txscript.OP_PUSHDATA4:
		size = int(binary.LittleEndian.Uint32(sizeBytes))
	}
	*posPointer += size
	return size
}

func getContentLength(raw []byte, pos *int) int {
	total := 0
	length := getPushDataLength(raw, pos)
	for length > 0 {
		total += length
		opcode := raw[*pos]
		if opcode == txscript.OP_ENDIF {
			break
		} else if opcode == txscript.OP_0 {
			// 某些情况会用OP_0分割，跳过
			*pos++
		}
		length = getPushDataLength(raw, pos)
	}
	return total
}

func ParseInscription(txWitness [][]byte) ([]map[int][]byte, error) {
	// 规则：一个信封，就是一次铭刻。
	// 无效情况：1. 存在不支持的指令；2.信封内部嵌套信封
	// 可能存在任何一个witness

	result := make([]map[int][]byte, 0)
	for _, raw := range txWitness {

		pos := int(0)
		for pos < len(raw) {
			begin := getBeginPosition(raw[pos:])
			if begin < 0 {
				break
			}
			begin += pos

			end := getEndPosition(raw[begin:])
			if end < 0 {
				break
			}
			end += begin
			pos = end + 1

			envelope := raw[begin:pos]

			fields := make(map[int][]byte)
			length := end - begin
			i := 0
			for i < length {
				opcode := envelope[i]
				if opcode == txscript.OP_0 {
					// body
					i++
					content, err := readContent(envelope, &i)
					if err != nil {
						break
					}
					fields[FIELD_CONTENT] = content
				} else if opcode == txscript.OP_ENDIF {
					i++
					break
				} else {
					// read tags
					i++
					tagType := readPushData(envelope, &i, opcode)
					if tagType == nil {
						continue
					} else {
						opcode = envelope[i]
						i++
						tagContent := readPushData(envelope, &i, opcode)

						if len(tagType) == 1 {
							fields[int(tagType[0])] = tagContent
						} else {
							if tagContent == nil {
								fields[FIELD_INVALID1] = tagType
							} else if len(tagContent) == 1 {
								fields[int(tagContent[0])] = tagType
							} else {
								fields[FIELD_INVALID1] = tagType
								fields[FIELD_INVALID2] = tagContent
							}
						}
					}
				}
			}
			result = append(result, fields)
		}

	}

	return result, nil
}

func ParseBrc20Content(content string) *Brc20BaseContent {
	var ret Brc20BaseContent
	err := json.Unmarshal([]byte(content), &ret)
	if err != nil {
		Log.Warnf("invalid json: %s, %v", content, err)
		return nil
	}
	ret.Ticker = strings.TrimSpace(ret.Ticker)
	return &ret
}

func Cbor2json(cborData []byte) ([]byte, error) {
	if cborData == nil {
		return nil, fmt.Errorf("no data")
	}
	var decodedData map[string]string
	err := cbor.Unmarshal(cborData, &decodedData)
	if err != nil {
		return nil, err
	}
	jsonData, err := json.Marshal(decodedData)
	if err != nil {
		return nil, err
	}
	return (jsonData), nil
}

func Json2cbor(jsonData []byte) ([]byte, error) {
	if jsonData == nil {
		return nil, fmt.Errorf("no data")
	}
	var decodedData map[string]string
	err := json.Unmarshal(jsonData, &decodedData)
	if err != nil {
		return nil, err
	}

	cborData, err := cbor.Marshal(decodedData)
	if err != nil {
		return nil, err
	}
	return (cborData), nil
}

func GetProtocol(fields map[int][]byte) (string, []byte) {
	content := (fields)[FIELD_CONTENT]
	protocol, ok := (fields)[FIELD_META_PROTOCOL]
	if ok {
		jsonStr, err := Cbor2json((fields)[FIELD_META_DATA])
		if err == nil {
			content = jsonStr
		}
		return string(protocol), content
	}

	var ordxContent OrdxBaseContent
	err := json.Unmarshal([]byte(content), &ordxContent)
	if err != nil {
		return "", nil
	}

	return ordxContent.P, content
}

func IsValidName(name string) bool {
	// 使用正则表达式匹配标点符号/空格/控制符
	if name == "" {
		return false
	}
	reg := regexp.MustCompile(`[\pP\pZ\pC]`)
	return !reg.MatchString(name)
}

func IsValidSat20Name(name string) bool {
	return IsValidName(name) && IsValidNameLen(name)
}

func IsValidNameLen(name string) bool {
	tickLen := len(name)
	return (tickLen >= MIN_NAME_LEN && tickLen <= MAX_NAME_LEN)
}

func PreprocessName(name string) string {
	return strings.TrimSpace(name)
}

func IsValidSNSName(name string) bool {
	if len(name) > MAX_NAME_LEN {
		return false
	}
	parts := strings.Split(name, ".")
	l := len(parts)
	bReg := false
	if l == 1 {
		bReg = IsValidSat20Name(parts[0])
	} else if l == 2 {
		bReg = IsValidName(parts[0]) && IsValidName(parts[1])
	}
	return bReg
}
