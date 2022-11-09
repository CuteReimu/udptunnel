package main

import (
	"flag"
	"fmt"
	"github.com/davyxu/cellnet/msglog"
	"io"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/davyxu/golog"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

var tunnelPort = flag.Int64("p", 12345, "tunnel port")
var isVpnServer = flag.Bool("s", false, "true means is a vpn server")

func main() {
	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}
	if *isVpnServer {
		(&vpnServer{timeout: time.Minute}).start()
		return
	}
	answer := input("你是作为服务器还是客户端？1.服务器，2.客户端", func(ret int) bool { return ret == 1 || ret == 2 })
	buf, _ := os.ReadFile("ip.conf")
	var address string
	if len(buf) == 0 {
		address = input[string]("请输入连接的地址，例如 10.8.0.1")
	} else {
		address = input[string]("你上次连接的地址是" + string(buf) + "，你可以直接回车以连接这个地址，或者输入新的地址：")
		if len(address) == 0 {
			address = string(buf)
		}
	}
	if address != string(buf) {
		_ = os.WriteFile("ip.conf", []byte(address), 0666)
	}
	if answer == 1 {
		port := input("请输入目标程序监听的端口，例如 5230", func(ret uint32) bool { return ret <= 65535 })
		(&server{port: port, timeout: time.Minute}).start(address)
	} else {
		(&client{timeout: time.Minute}).start(address)
	}
}

func input[T any](question string, check ...func(ret T) bool) (ret T) {
	for {
		fmt.Println(question)
		n, err := fmt.Scanln(&ret)
		if reflect.TypeOf(ret).Kind() == reflect.String || err == nil && n == 1 && (len(check) == 0 || check[0](ret)) {
			return
		}
		fmt.Println("输入有误，请重新输入！")
	}
}

var log = golog.New("tunnel")

func init() {
	msglog.SetCurrMsgLogMode(msglog.MsgLogMode_Mute)
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
