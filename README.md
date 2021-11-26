# Shuttle
[![ci-test-build](https://github.com/cyejing/shuttle/actions/workflows/ci-test-build.yml/badge.svg)](https://github.com/cyejing/shuttle/actions/workflows/ci-test-build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/cyejing/shuttle)](https://goreportcard.com/report/github.com/cyejing/shuttle)

Shuttle是功能小巧的代理网关，连通每个地方，互联万物。

## 使用
下载执行文件[Release页面](https://github.com/cyejing/shuttle/releases)
### 启动服务端
``./shuttles -c example/shuttles.yaml``

配置参数
```yaml
#example/shuttles.yaml
addr: 127.0.0.1:4880  #http端口
sslAddr: 127.0.0.1:4843 #https端口
cert: example/s.cyejing.cn_chain.crt #https证书
key: example/s.cyejing.cn_key.key #https证书
passwords:
  - cyejing123 #对应客户端密码
routes: #网关路由
  - id: na0mdwfr0lfuv4rubvt4gsg805uofhhk
    order: 100
    host: .* #正则匹配域名
    loggable: true
    filters:
      - name: resource #本地静态资源
        params:
          root: "./html"
```
### 启动客户端
``./shuttlec -c example/shuttlec.yaml``

配置参数
```yaml
runType: socks #运行类型socks 代理
localAddr: 127.0.0.1:1080 #本地socks5代理
remoteAddr: 127.0.0.1:4843 #服务器地址
password: cyejing123 #对应服务器密码

```

### 浏览器设置socks5代理
Enjoy
