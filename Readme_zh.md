# idgo 简介
[![Build Status](https://travis-ci.org/flike/idgo.svg?branch=master)](https://travis-ci.org/flike/idgo)
## 1. idgo特点

idgo是一个利用MySQL批量生成ID的ID生成器, 主要有以下特点:

- 生成的ID是顺序递增的。
- 每次通过事务批量取ID,性能较高,且不会对MySQL造成压力。
- 当ID生成器服务崩溃后,可以继续生成有效ID,避免了ID回绕的风险。
- 服务端模拟Redis协议，通过`GET`和`SET`获取和设置key。不必开发专门的获取ID的SDK，直接使用Reids的SDK就可。

业界已经有利于MySQL生成ID的方案,都是通过:

```
REPLACE INTO Tickets64 (stub) VALUES ('a');
SELECT LAST_INSERT_ID();
```
这种方式生成ID的弊端就是每生成一个ID都需要查询一下MySQL,当ID生成过快时会对MySQL造成很大的压力。这正式我开发这个项目的原因。服务端兼容Redis协议是为了避免单独开发和idgo通信的SDK。

## 2. idgo架构
idgo和前端应用是采用redis协议通信的，然后每个`id_key`是存储在MySQL数据库中，每个key会在MySQL中生成一张表，表中只有一条记录。这样做的目的是保证当idgo由于意外崩溃后，`id_key`对应的值不会丢失，这样就避免产生了id回绕的风险。
![idgo_arch](http://ww2.sinaimg.cn/large/6e5705a5gw1f2nz3bot3tj20qo0k0mxe.jpg)

idgo目前只支持四个redis命令：

```
1. SET key value,通过这个操作设置id生成器的初始值。
例如：SET abc 123
2. GET key,通过该命令获取id。
3. EXISTS key,查看一个key是否存在。
4. DEL key,删除一个key。
5. SELECT index,选择一个db，目前是一个假方法，没实现任何功能，只是为了避免初始化客户端时调用SELECT出错。
```


## 3. 安装和使用idgo

1. 安装idgo

```
	1. 安装Go语言环境（Go版本1.3以上），具体步骤请Google。
	2. 安装godep工具, go get github.com/tools/godep 。 
	2. git clone https://github.com/flike/idgo src/github.com/flike/idgo
	3. cd src/github.com/flike/idgo
	4. source ./dev.sh
	5. make
	6. 设置配置文件
	7. 运行idgo。./bin/idgo -config=etc/idgo.toml
```


设置配置文件(`etc/idgo.toml`):

```
#idgo的IP和port
addr="127.0.0.1:6389"
#log_path: /Users/flike/src 
#日志级别
log_level="debug"

[storage_db]
mysql_host="127.0.0.1"
mysql_port=3306
db_name="idgo_test"
user="root"
password=""
max_idle_conns=64
```

操作演示：

```
#启动idgo
➜  idgo git:(master) ✗ ./bin/idgo -config=etc/idgo.toml
2016/04/07 11:51:20 - INFO - server.go:[62] - [server] "NewServer" "Server running" "netProto=tcp|address=127.0.0.1:6389" req_id=0
2016/04/07 11:51:20 - INFO - main.go:[80] - [main] "main" "Idgo start!" "" req_id=0

#启动一个客户端连接idgo
➜  ~  redis-cli -p 6389
redis 127.0.0.1:6389> get abc
(integer) 0
redis 127.0.0.1:6389> set abc 100
OK
redis 127.0.0.1:6389> get abc
(integer) 101
redis 127.0.0.1:6389> get abc
(integer) 102
redis 127.0.0.1:6389> get abc
(integer) 103
redis 127.0.0.1:6389> get abc
(integer) 104
redis 127.0.0.1:6389>

```

## 4. 压力测试
压测环境

|类别|名称|
|---|---|
|OS       |CentOS release 6.4|
|CPU      |Common KVM CPU @ 2.13GHz|
|RAM      |2GB|
|DISK     |50GB|
|Mysql    |5.1.73|

本地mac连接远程该MySQL实例压测ID生成服务。
每秒中可以生成20多万个ID。性能方面完全不会有瓶颈。

## 5.ID生成服务宕机后的恢复方案

当idgo服务意外宕机后，可以切从库，然后将idgo对应的key加上适当的偏移量。

# License

MIT
