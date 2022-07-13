# 保护mc服务器免受ddos的侵扰

对于ddos的发起者技术水平我们将其从低到高分为

1. 只会打开客户端查看域名然后购买ddos。dns使用srv记录而不是a记录即可有效防御
2. 了解dns解析基本原理或者懂得使用资源监视器查看进程创建的连接。使用srv记录无法防御此类攻击者，需要购买高防ip或购买cdn使用http隧道防御
3. 会使用wireshark进行简单的抓包。购买高防ip防御或购买cdn使用https隧道防御
4. 了解https基本原理。需要购买高防ip或购买cdn使用伪装sni的https隧道防御
5. 会安装证书解密https流量。需要购买高防ip或购买cdn使用开启双向认证的的伪装sni的https隧道进行防御(还未实现)
6. 懂得反编译和逆向相关知识。~~加钱买高防~~

选择cdn之前要确认其支持**域前置**

假设我们使用百度智能云的cdn
构建前需确认已经安装golang最新版并将go添加进环境变量，windows添加环境变量之后需重启

执行

```
git clone https://github.com/yu1745/websocket_tunnel.git
cd websocket_tunnel
./build.ps1
```

此时生成的server.exe可用，但是client.exe不可用
考虑到client.exe要随mc客户端一起分发，使用命令行传参的方式调用就有泄露域名的风险，因此将参数在编译时写死是更好的选择。
打开build.ps1，需要为client修改的参数有5个

| 参数名 | 默认值 | 可选值 | 例子 | 意义|
| --- | --- | --- | --- | --- |
| addr | 无 | 域名+**英文**冒号+端口号 | opencdn.jomodns.com:443 | 要连接的cdn节点的cname地址，需要用户手动寻找 [如何寻找(还未完工)](https://baidu.com)
| https | false | true\|false | true | 是否开启https，不开启https有暴露域名的风险
| listen | localhost:25565 | localhost+英文冒号+可监听的端口号 | localhost:9999 | mc客户端连接的地址 |
| fake | 无 | 任意合法域名(仅域名，不需要端口号) | www.baidu.com | 与cdn建立https连接时发送的sni |
| real | 无 | 任意合法域名(仅域名，不需要端口号) | mc.example.com | 用户在cdn处登记的需要加速的域名
