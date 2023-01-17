# UDP Tunnel

![](https://img.shields.io/github/go-mod/go-version/CuteReimu/udptunnel "语言")
[![](https://img.shields.io/github/actions/workflow/status/CuteReimu/udptunnel/golangci-lint.yml?branch=master)](https://github.com/CuteReimu/udptunnel/actions/workflows/golangci-lint.yml "代码分析")
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

## 举例

我们用“流星蝴蝶剑”联机为例。众所周知，“流星蝴蝶剑”是使用的udp进行联机的，房主监听的是udp端口5230。如下步骤进行即可：

1. 首先，你需要一个有公网ip的机器，例如XX云服务器，在上面运行`udptunnel.exe -s`。**别忘了设置服务器的防火墙，将udp端口12345设置为允许**。
2. 房主在自己的电脑上运行`udptunnel.exe`，选择服务器，然后端口输入5230。之后房主就可以进入游戏建房了。
3. 其它玩家在自己的电脑上运行`udptunnel.exe`，选择客户端，然后选择房主的服务器。之后进行游戏，搜索房间列表，就能搜到房间了。

说明：

1. 很有可能，你的XX云服务器是linux机器，而自己的电脑是windows机器，那么你需要编译出两个系统版本的udptunnel。
2. 如果不想监听12345端口，可以运行`udptunnel.exe -s -p 12345`自行指定端口，那么所有人连接时也要同样输入`udptunnel.exe -p 12345`。
