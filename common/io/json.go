package io

import "encoding/json"

func ToByte(message interface{}) (marshal []byte, err error) {
	marshal, err = json.Marshal(message)
	return
}

func ToObj(marshal []byte, message interface{}) (err error) {
	err = json.Unmarshal(marshal, message)
	return
}
