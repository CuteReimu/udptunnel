# UDP Tunnel

![](https://img.shields.io/github/languages/top/CuteReimu/udptunnel "语言")
[![](https://img.shields.io/github/workflow/status/CuteReimu/udptunnel/Go)](https://github.com/CuteReimu/udptunnel/actions/workflows/golangci-lint.yml "代码分析")
[![](https://img.shields.io/github/contributors/CuteReimu/udptunnel)](https://github.com/CuteReimu/udptunnel/graphs/contributors "贡献者")
[![](https://img.shields.io/github/license/CuteReimu/udptunnel)](https://github.com/CuteReimu/udptunnel/blob/master/LICENSE "许可协议")

考虑到联机时用udp进行广播搜索房间时，OpenVPN无法转发udp消息，因此做了一个基于kcp的tunnel，用于转发udp消息

## 使用方法

```bash
go build -o udptunnel.exe
```

然后双击运行生成出来的`udptunnel.exe`即可

首先，你需要一个有公网ip的机器，在上面运行`udptunnel.exe -s`

然后建房间的玩家启动`udptunnel.exe`之后选择服务器，其它玩家选择客户端。之后再在游戏中创建房间，就能正常搜索房间了
