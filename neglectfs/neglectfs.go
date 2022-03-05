package neglectfs

func GetNeglectFileNames() ([]string, *ReadLine, error) {

	// 默认忽略文件
	fileConfigPath := "./.ignore-files"

	// 默认忽略的文件或目录的内容
	defaultNeglectFileNames := []string{
		"vendor",
		".idea",
		"runtime",
		".git",
		".DS_Store",
		".gitignore",
		".ignore-files",
	}

	// 行写入
	fileLine, err := GetInstance(fileConfigPath)
	if err != nil {
		return nil, fileLine, err
	}

	// 行读取
	files := []string{}
	if fileLine.IsExist() {
		if _, err := fileLine.WriteLine(defaultNeglectFileNames); err != nil {
			return nil, fileLine, err
		}
	}
	_, err = fileLine.ReadLine(&files, nil)
	if err != nil {
		panic(err)
	}

	return files, fileLine, nil
}
