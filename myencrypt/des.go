package myencrypt

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"fmt"
	"strconv"
)

// DES加密
// iv为空则采用ECB模式，否则采用CBC模式
func DesEncrypt(value, secretKey, iv string) (string, error) {
	if value == "" {
		return "", nil
	}

	//根据秘钥生成8位的秘钥切片
	keyBytes := make([]byte, des.BlockSize)
	copy(keyBytes, []byte(secretKey))
	//获取block
	block, err := des.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	blocksize := block.BlockSize()
	valueBytes := []byte(value)

	//填充
	fillsize := blocksize - len(valueBytes)%blocksize
	repeat := bytes.Repeat([]byte{byte(fillsize)}, fillsize)
	valueBytes = append(valueBytes, repeat...)

	result := make([]byte, len(valueBytes))

	//加密
	if iv == "" {
		temp := result
		for len(valueBytes) > 0 {
			block.Encrypt(temp, valueBytes[:blocksize])
			valueBytes = valueBytes[blocksize:]
			temp = temp[blocksize:]
		}
	} else {
		//向量切片
		ivBytes := make([]byte, des.BlockSize)
		copy(ivBytes, []byte(iv))

		encrypter := cipher.NewCBCEncrypter(block, ivBytes)
		encrypter.CryptBlocks(result, valueBytes)
	}

	//以hex格式数值输出
	encryptText := fmt.Sprintf("%x", result)
	return encryptText, nil
}

// DES解密
// iv为空则采用ECB模式，否则采用CBC模式
func DesDecrypt(value, secretKey, iv string) (string, error) {
	if value == "" {
		return "", nil
	}

	//根据秘钥生成8位的秘钥切片
	keyBytes := make([]byte, des.BlockSize)
	copy(keyBytes, []byte(secretKey))
	//获取block
	block, err := des.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	//将hex格式数据转换为byte切片
	valueBytes := []byte(value)
	var encryptedData = make([]byte, len(valueBytes)/2)
	for i := 0; i < len(encryptedData); i++ {
		b, err := strconv.ParseInt(value[i*2:i*2+2], 16, 10)
		if err != nil {
			return "", err
		}
		encryptedData[i] = byte(b)
	}

	result := make([]byte, len(encryptedData))

	if iv == "" {
		blocksize := block.BlockSize()
		temp := result
		for len(encryptedData) > 0 {
			block.Decrypt(temp, encryptedData[:blocksize])
			encryptedData = encryptedData[blocksize:]
			temp = temp[blocksize:]
		}
	} else {
		//向量切片
		ivBytes := make([]byte, des.BlockSize)
		copy(ivBytes, []byte(iv))

		//解密
		blockMode := cipher.NewCBCDecrypter(block, ivBytes)
		blockMode.CryptBlocks(result, encryptedData)
	}

	//取消填充
	unpadding := int(result[len(result)-1])
	result = result[:(len(result) - unpadding)]
	return string(result), nil
}
