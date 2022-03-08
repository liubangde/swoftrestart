package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"swoftrestart/neglectfs"

	"github.com/fsnotify/fsnotify"
)

func main() {
	// 获取命令行命令
	var targetInstruction string
	flag.StringVar(&targetInstruction, "t", "php ./bin/swoft http:start", "需要执行的命令")
	flag.Parse()

	// 获取需要监听的文件
	watcherFileNames := []string{}
	f := getFileList("./", &watcherFileNames)
	defer f.Close()

	// 获取当前所在目录
	str, _ := os.Getwd()

	// 启动进程
	//run := make(chan struct{}, 1)
	stop := make(chan struct{}, 1)
	quit := make(chan int, 1)
	log.Println("监听当前所在目录=========== ", str)
	go start(quit, targetInstruction)

	// 创建监听
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// 监听推出管道
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					// 进行热重启-
					log.Println("监听到文件修改了，进行进程重启.......", event.Name)
					stop <- struct{}{}
					stop_(stop, quit)
					go start(quit, targetInstruction)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("监听错误error:", err)
			}
		}
	}()

	// 监听文件
	for _, fileName := range watcherFileNames {
		err = watcher.Add(fileName)
	}
	if err != nil {
		log.Fatal("错误信息：", err)
	}

	<-done
}

// 启动进程
func start(quit chan int, targetInstruction string) {
	cmd := exec.Command("/bin/bash", "-c", targetInstruction)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// start
	log.Println("启动swoft.......", cmd.Args)
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	// 获取进程号
	log.Println("pid 是：.", cmd.Process.Pid)
	quit <- cmd.Process.Pid
}

func stop_(stop chan struct{}, quit chan int) {
	pid, _ := <-quit

	log.Println("结束进程：", pid)
	cmd := exec.Command("/bin/bash", "-c", "/bin/kill -9 "+strconv.Itoa(pid))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalf("命令错误: %v", err)
	}

	<-stop
}

// 获取文件列表
func getFileList(dir string, watcherFileNames *[]string) *neglectfs.ReadLine {

	neglectFiles, f, err := neglectfs.GetNeglectFileNames()
	if err != nil {
		panic(err)
	}

	//切片转Map
	neglectFileMap := make(map[string]struct{})
	for _, neglectFile := range neglectFiles {
		neglectFileMap[neglectFile] = struct{}{}
	}

	// 获取当前目录下的所有文件
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {

		// 监听文件筛选
		if _, ok := neglectFileMap[f.Name()]; ok {
			continue
		}

		// 是目录?
		var fileName bytes.Buffer
		fileName.WriteString(dir)
		fileName.WriteString(f.Name())

		s, err := os.Stat(fileName.String())
		if err != nil {
			log.Println("解析文件失败!", fileName.String(), err.Error())
		}
		if !s.IsDir() {
			*watcherFileNames = append(*watcherFileNames, fileName.String())
			continue
		}

		// 下轮扫描
		fileName.WriteString("/")
		getFileList(fileName.String(), watcherFileNames)
	}

	return f
}
