package main

import (
  "bytes"
  "encoding/binary"
  "log"
  "math/big"
  "os"
	"encoding/gob"
  "reflect"
)

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
var base = big.NewInt(int64(len(b58Alphabet)))

func IntToHex(num int64) []byte {
  buff := new(bytes.Buffer)
  err := binary.Write(buff, binary.BigEndian, num)
  if err != nil {
    log.Panic(err)
  }
  return buff.Bytes()
}

func ReverseBytes(data []byte) {
  for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
    data[i], data[j] = data[j], data[i]
  }
}

func Base58Encode(data []byte) []byte{
  var result []byte
  x := big.NewInt(0).SetBytes(data)
  mod := &big.Int{}
  zero := big.NewInt(0)

  for x.Cmp(zero) != 0 {
    x.DivMod(x, base, mod)
    result = append(result, b58Alphabet[mod.Int64()])
  }
  // https://en.bitcoin.it/wiki/Base58Check_encoding#Version_bytes
  if data[0] == 0x00 {
    result = append(result, b58Alphabet[0])
  }

  ReverseBytes(result)
  return result
}

func Base58Decode(data []byte) []byte{
  result := big.NewInt(0)
  for _, b := range data {
    result.Mul(result, base)
    result.Add(result, big.NewInt(int64(bytes.IndexByte(b58Alphabet, b))))
  }
  decoded := result.Bytes()
  if data[0] == b58Alphabet[0] {
    decoded = append([]byte{0x00}, decoded...)
  }
  return decoded
}

func IsFileExist(file string) bool{
  _, err := os.Stat(file)
  return !os.IsNotExist(err)
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	err := gob.NewEncoder(&buff).Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func gobDecode(data [] byte, e interface{}) {
  err := gob.NewDecoder(bytes.NewReader(data)).DecodeValue(reflect.ValueOf(e))
  if err != nil {
    log.Panic(err)
  }
}

func stringInSlice(item string, list []string) bool {
  for _, l := range list {
    if l == item {
      return true
    }
  }
  return false
}
