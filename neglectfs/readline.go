package neglectfs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type ReadLine struct {
	file   *os.File
	reader *bufio.Reader
	path   string
}

var instance *ReadLine

func GetInstance(filePath string) (*ReadLine, error) {
	if instance == nil {
		instance = &ReadLine{
			path: filePath,
		}
	}
	return instance, nil
}

// 文件是否存在
func (r *ReadLine) IsExist() bool {
	_, err := os.Lstat(r.path)
	return os.IsNotExist(err)
}

// 支持切片||管道存放数据内容
func (r *ReadLine) ReadLine(arr *[]string, data chan string) (*ReadLine, error) {

	// 是否有文件依据 ?
	if r.file == nil {
		var err error
		r.file, err = os.Open(r.path)
		if err != nil {
			panic(err)
		}
	}

	// 当前指针复位
	if currentPointer, err := r.file.Seek(0, 1); err != nil {
		panic(err)
	} else {
		if currentPointer != 0 {
			r.file.Seek(0, 0)
		}
	}

	// 创建文件缓存区
	r.reader = bufio.NewReader(r.file)

	for {
		//行读取
		lineContent, err := r.reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			panic(err)
		}
		if len(lineContent) == 0 {
			break
		}

		// 保存结果
		lineText := strings.Replace(
			strings.Replace(string(lineContent), "\n", "", -1), "\r", "", -1)
		if arr != nil {
			*arr = append(*arr, lineText)
		} else {
			data <- lineText
		}

		// 监听信号量判断是否完成 ?
		if err == io.EOF {
			// 如果是管道? 关闭管道
			if data != nil {
				close(data)
			}
			break
		}
	}

	return r, nil
}

func (r *ReadLine) SetPath(path string) {
	r.path = path
}

func (r *ReadLine) WriteLine(fileNames []string) (*ReadLine, error) {

	// 是否存在忽略文件
	if r.IsExist() {
		var err error
		r.file, err = os.Create(r.path)
		if err != nil {
			return r, err
		}
	}

	// 创建并写入默认忽略目录
	b := bufio.NewWriter(bufio.NewWriter(r.file))
	for _, v := range fileNames {
		_, err := fmt.Fprintln(b, v)
		if err != nil {
			panic(err)
		}
	}

	b.Flush()
	return r, nil
}

func (r *ReadLine) Close() {
	r.file.Close()
}
