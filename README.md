# Shuttle
[![ci-test-build](https://github.com/cyejing/shuttle-go/actions/workflows/ci-test-build.yml/badge.svg)](https://github.com/cyejing/shuttle-go/actions/workflows/ci-test-build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/cyejing/shuttle)](https://goreportcard.com/report/github.com/cyejing/shuttle)

Shuttle让互联网更简单

## Feature

- tls加密通信通道
- 内网穿透, 支持动态调节

## Architecture

![architecture](/doc/pic/architecture.png)

## Download
下载可执行文件[Release页面](https://github.com/cyejing/shuttle-go/releases)

## Quick Start 

### Socks5代理使用
#### Start Server
``./shuttles -c example/shuttles.yaml``

配置参数
```yaml
#example/shuttles.yaml
addrs:
  - addr: 127.0.0.1:4880
  - addr: 127.0.0.1:4843
    cert: example/chain.crt #https证书
    key: example/key.key #https证书
trojan:
  passwords:
    - sQtfRnfhcNoZYZh1wY9u #对应客户端密码
```
#### Start Client
``./shuttlec -c example/shuttlec-socks.yaml``

配置参数
```yaml
runType: socks #运行类型socks 代理
localAddr: 127.0.0.1:1080 #本地socks5代理
remoteAddr: 127.0.0.1:4843 #服务器地址
password: sQtfRnfhcNoZYZh1wY9u #对应服务器密码

```

#### 浏览器设置socks5代理
Enjoy

### Wormhole穿透使用
#### Start Server
``./shuttles -c example/shuttles.yaml``

配置参数
```yaml
#example/shuttles.yaml
addrs:
  - addr: 127.0.0.1:4880
wormhole:
  passwords:
    - 58JCEmvcBkRAk1XkK1iH
```
#### Start Client
``./shuttlec -c example/shuttlec-wormhole.yaml``

配置参数
```yaml
runType: wormhole
name: unique-name
sslEnable: false
remoteAddr: 127.0.0.1:4880
password: 58JCEmvcBkRAk1XkK1iH

ships:
  - name: test
    remoteAddr: 127.0.0.1:4022
    localAddr: 127.0.0.1:22

```

#### Enjoy Internet
ship-tcp -> remoteAddr -> localAddr

### Route代理使用
#### Start Server
``./shuttles -c example/shuttles.yaml``

配置参数
```yaml
#example/shuttles.yaml
addrs:
  - addr: 127.0.0.1:4880
  - addr: 127.0.0.1:4843
    cert: example/chain.crt #https证书
    key: example/key.key #https证书
gateway:
  routes:
    - id: APUGW4UDKHgRX8bQuqRErTn9LGwyuFfV
      order: 100
      host: .* #正则匹配域名
      loggable: true
      filters:
        - name: resource
          params:
            root: "./html"
    - id: L28dECFtGfGfP2BTN9iNvkUEm2BWLMw9
      order: 120
      path: /proxy  #正则匹配路径
      loggable: true
      filters:
        - name: rewrite
          params:
            regex: "/proxy/(.*)"
            replacement: "/$1"
        - name: proxy
          params:
            uri: "http://127.0.0.1:8088"
```
