翻墙DNS fqdns
=====

# Introduction
Distribute to different DNS server based on request domain names
根据域名列表转发到不同的dns服务器以获取未被污染的ip
##举个栗子：
有一个小团队大概20多号人吧，在路由器后面上网，现在有科学上网的需求了。每个人都代理上网虽然可行，但是不是每个人都能弄正确。而且遇到是应用需要科学上网的，又不支持代理什么的，就悲剧了。所以就考虑在路由器上翻墙。但是如果全部流量都走路由器的vpn出口，那么看个优酷肯定是卡的，而且流量很快就用光。于是考虑到国内国外的ip分开路由，这个事情简单，国内国外的ip段分开做route。那么问题来了，路由是通过ip来实现的，如果域名已经被污染，解析出来的是错误的ip，那么这个路由方案就是不完整的，不能访问的还是不能访问。于是就有了这么一个程序，帮助我们获得真实的ip。

# 预编译包下载地址
[百度云盘](http://pan.baidu.com/s/1pJ8x0rL)

# 原理
* fqdns维护一个国内域名的列表,在这个列表中的域名通过国内的dns解析,确保正确的能区分电信/网通/cdn
* 不在列表里面的域名将会发送到远端服务器解析,比如opendns,但是如果在不通过vpn路由的情况下,数据包依然可能会被污染
* fqns可以配置使用tcp的方式和远端的fqdns服务器通讯,让远端的fqns服务器解析完成之后通过tcp返回给国内的dns分发服务器,绕开墙的检查
* 考虑到有很多网站是使用cdn做分发,fqdns有一个白名单和黑名单,白名单的域名一定会使用国内dns解析,比如各种cdn;黑名单的域名一定使用远端解析,比如google
这样可以确保能得到一个正确的解析结果

# 结构
* fqdns可以工作在两个模式下面,一个是disp分发模式,以这个模式启动后,所有的dns请求会根据域名列表分发到对应的dns服务器上,比如一个国内,一个国外
* 另外一个是resolver解析模式,这个模式启动的服务器将会等待其他服务器发送的tcp请求,然后返回tcp结果
* 标准流程如下: 用户系统发出dns请求->fqdns收到标准dns请求后判断是否国内域名->国内域名使用国内dns解析后返回->国外域名通过tcp转发到远端fqdns上解析
* disp模式的fqdns可以在本地启动,局域网机器上启动,也可以在公网服务器上启动,然后通过tcp和国外的fqdns通讯

# 场景
fqdns的设计目的是尽可能的绕开GFW的污染,获取能真实使用的ip,有部分应用并不原生支持socks5代理,dns解析会使用本地的dns服务器,如果出现被污染的情况将会无法使用,比如dropbox的installer.

# 编译和依赖
golang
[golang.org](http://golang.org)

golang dns库
[golang dns lib](https://github.com/miekg/dns.git)
安装方法
> go get github.com/miekg/dns

# 配置
fqdns有1个配置文件和3个列表文件
## config 配置文示例
disp 模式
```
{
	"local":["114.114.114.114:53"],
	"remote":["127.0.0.1:37241"],
	"port":53,
	"pac":"/Users/yourname/fqdns/whitelist.pac",
	"white":"/Users/yourname/fqdns/white",
	"black":"/Users/yourname/fqdns/black",
	"tcpremote":true
}
```
* local 国内的dns服务器,
* remote 转发的远端dns服务器
* port 本地监听的端口,默认是53
* pac 文件来自于这里[thanks breakwa11](https://github.com/breakwa11/gfw_whitelist/blob/master/whitelist.pac)
* white 白名单文件
* black 黑名单文件
* tcpremote 是否使用tcp方式请求远端remote服务器,false为使用标准dns的udp方式


resolver模式
```
{
	"local":[],
	"remote":["8.8.8.8:53"],
	"port":37241,
	"white":"",
	"tcpremote":false
}
```
* local 未使用
* remote 真实解析的dns服务器
* port 监听的tcp端口,对应disp模式配置中的remote
* white 未使用
* black 未使用
* pac  未使用
* tcpremote 未使用

## 黑名单白名单
### white
```
#domains in this will goes to GFWed/local dns ie. 114.114.114.114
#expected cdn domains need to use local dns to dispatch nearest servers
#one domain pattern per line, support wildcards match ie. *.cloudfront.com
*.akamai.net
*.cloudfront.net
cdn.acewebgames.com
*.cn
```
### black
```
#domains in this file will goes to remote resolver
#use to make sure domains resolved outside GFW, even it is in whitelist.pac
#wildcards match supported
*.google.com
*.googleapi.com
*.google.com.*
*.googlecode.com
google.com

```

# 启动方式
`fqdns -config config.json -mode disp`

# 客户端配置
修改dns服务器地址为disp模式的fqdns

