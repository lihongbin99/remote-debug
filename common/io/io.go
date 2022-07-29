package io

import (
	"fmt"
	"net"
	"remote-debug/common/utils"
)

// SendMessage TODO 加密消息
func SendMessage(conn *net.TCPConn, message interface{}) error {
	// 解析
	data, err := ToByte(message)
	if err != nil {
		return err
	} else if len(data) <= 0 {
		return nil
	}

	return SendData(conn, data)
}

func ReadMessage(conn *net.TCPConn, message interface{}) error {
	// 读取数据
	data, err := ReadData(conn)
	if err != nil {
		return err
	}

	// 解析
	if err = ToObj(data, message); err != nil {
		return err
	}

	return nil
}

func SendData(conn *net.TCPConn, data []byte) error {
	// 发送消息长度
	if _, err := conn.Write(utils.I2b32(uint32(len(data)))); err != nil {
		return err
	}
	// 发送消息
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

func ReadData(conn *net.TCPConn) ([]byte, error) {
	// 读取前缀
	buf := make([]byte, 4)
	readSum := 0
	for readSum < len(buf) {
		if readLength, err := conn.Read(buf[readSum:]); err != nil {
			return nil, err
		} else {
			readSum += readLength
		}
	}

	// 获取消息长度
	messageLen32, err := utils.B2i32(buf)
	if err != nil {
		return nil, err
	}
	messageLen := int(messageLen32)
	if messageLen <= 0 {
		return nil, fmt.Errorf("message len error: %d", messageLen)
	}

	// 读取消息
	data := make([]byte, messageLen)
	readSum = 0
	for readSum < messageLen {
		if readLength, err := conn.Read(data[readSum:]); err != nil {
			return nil, err
		} else {
			readSum += readLength
		}
	}

	return data, nil
}
