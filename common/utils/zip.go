package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func Zip(baseDir string, ignore string) ([]byte, error) {
	dir, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	// 解析忽略
	ignoreMap := make(map[string]int8)
	ignores := strings.Split(ignore, ",")
	for _, i := range ignores {
		ignoreMap[i] = 1
	}

	buffer := bytes.NewBuffer([]byte{})
	zipWriter := zip.NewWriter(buffer)

	if err = doZip(dir, baseDir, "", zipWriter, ignoreMap); err != nil {
		return nil, err
	}
	_ = zipWriter.Close()
	return buffer.Bytes(), nil
}

func doZip(dir []os.DirEntry, baseDir, abDir string, zipWriter *zip.Writer, ignoreMap map[string]int8) error {
	for _, fileInfo := range dir {
		// 判断是否要跳过
		if _, ignore := ignoreMap[fileInfo.Name()]; ignore {
			continue
		}

		// 递归遍历所有文件夹
		if fileInfo.IsDir() {
			subDir, err := os.ReadDir(fmt.Sprintf("%s/%s/%s", baseDir, abDir, fileInfo.Name()))
			if err != nil {
				return err
			}
			if err = doZip(subDir, baseDir, fmt.Sprintf("%s%s/", abDir, fileInfo.Name()), zipWriter, ignoreMap); err != nil {
				return err
			}
			continue
		}

		// 读取文件
		file, err := os.Open(fmt.Sprintf("%s/%s%s", baseDir, abDir, fileInfo.Name()))
		if err != nil {
			return err
		}

		// 写入 zip
		writer, err := zipWriter.Create(fmt.Sprintf("%s%s", abDir, fileInfo.Name()))
		if err != nil {
			return err
		}
		if _, err = io.Copy(writer, file); err != nil {
			return err
		}

		// 关流
		_ = file.Close()
	}
	return nil
}

func UnZip(data []byte, outPath string) error {
	// 解析 zip
	reader := bytes.NewReader(data)
	zipFile, err := zip.NewReader(reader, int64(reader.Len()))
	if err != nil {
		return err
	}

	// 遍历 zip
	for _, file := range zipFile.File {
		fileNme := file.FileInfo().Name()
		abPath := file.Name[:len(file.Name)-len(fileNme)]

		// 创建文件夹
		err = os.MkdirAll(fmt.Sprintf("%s/%s", outPath, abPath), 0777)
		if err != nil {
			return err
		}

		// 读取文件
		readFile, err := file.Open()
		if err != nil {
			return err
		}

		// 写入文件
		writeFile, err := os.Create(fmt.Sprintf("%s/%s", outPath, file.Name))
		if err != nil {
			return err
		}
		if _, err = io.Copy(writeFile, readFile); err != nil {
			return err
		}

		// 关流
		_ = readFile.Close()
		_ = writeFile.Close()
	}
	return nil
}
