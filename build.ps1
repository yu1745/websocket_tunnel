$env:addr="example.com" #你的cdn的根cname 比如百度的就是opencdn.jomodns.com:80或者443
$env:https="false" #默认不开https 开启的话改为true
$env:listen=":25565" #游戏客户端连接端口
$env:fake= "other.com" #伪装域名 必须是与你使用同一家cdn的别人的域名 比如百度云的cdn就可以填www.zhujiba.com
$env:real= "" #你的真实域名 开启https时必填

cd client
go build -ldflags "-X main.addr=$env:addr -X main.https=$env:https -X main.listen=$env:listen -X main.fake=$env:fake -X main.real_=$env:real" client.go
mv client.exe ..
cd ..
cd server
go build server.go
mv server.exe ..
cd ..
