package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/davyxu/golog"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

var tunnelPort = flag.Int64("p", 12345, "tunnel port")

func main() {
	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}
	answer := input("你是作为服务器还是客户端？1.服务器，2.客户端", func(ret int) bool { return ret == 1 || ret == 2 })
	if answer == 1 {
		port := input("请输入目标程序监听的端口，例如 5230", func(ret int64) bool { return ret >= 0 && ret <= 65535 })
		(&server{port: port}).start()
	} else {
		address := input[string]("请输入连接的地址，例如 10.8.0.2")
		(&client{address: address}).start()
	}
}

func input[T any](question string, check ...func(ret T) bool) (ret T) {
	for {
		fmt.Println(question)
		n, err := fmt.Scanln(&ret)
		if err == nil && n == 1 && (len(check) == 0 || check[0](ret)) {
			return
		}
		fmt.Println("输入有误，请重新输入！")
	}
}

var log = golog.New("tunnel")

func init() {
	log.SetLevel(golog.Level_Debug)
	writer, err := rotatelogs.New(
		path.Join("log", "%Y-%m-%d.log"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		panic(err)
	}
	err = golog.SetOutput("tunnel", &logWriter{fileWriter: writer})
	if err != nil {
		return
	}
}

type logWriter struct {
	fileWriter io.Writer
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	_, _ = os.Stdout.Write(p)
	return l.fileWriter.Write(p)
}
