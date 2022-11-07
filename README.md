# UDP Tunnel

![](https://img.shields.io/github/languages/top/CuteReimu/udptunnel "语言")
[![](https://img.shields.io/github/workflow/status/CuteReimu/udptunnel/Go)](https://github.com/CuteReimu/udptunnel/actions/workflows/golangci-lint.yml "代码分析")
[![](https://img.shields.io/github/contributors/CuteReimu/udptunnel)](https://github.com/CuteReimu/udptunnel/graphs/contributors "贡献者")
[![](https://img.shields.io/github/license/CuteReimu/udptunnel)](https://github.com/CuteReimu/udptunnel/blob/master/LICENSE "许可协议")

考虑到联机时用udp进行广播搜索房间时，OpenVPN无法转发udp消息，因此做了一个tcp的tunnel，用于转发udp消息

## 使用方法

```bash
go build -o udptunnel.exe
```

然后双击运行生成出来的`udptunnel.exe`即可

可以使用 `udptunnel.exe -h` 获取详细参数信息

一方启动服务器，另一方启动客户端，之后再启动游戏创建房间，就能正常搜索房间了
