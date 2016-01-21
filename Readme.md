# idgo 简介

idgo是一个利用MySQL批量生成ID的ID生成器, 主要有以下特点:
- 每次通过事务批量取ID,性能较高,且不会对MySQL造成压力.
- 当ID生成器服务崩溃后,可以继续生成有效ID,避免了ID回绕的风险.

业界已经有利于MySQL生成ID的方案,都是通过:

```
REPLACE INTO Tickets64 (stub) VALUES ('a');
SELECT LAST_INSERT_ID();
```
这种方式生成ID的弊端就是每生成一个ID都需要查询一下MySQL,当ID生成过快时会对MySQL造成很大的压力.
这正是我写这个lib库的原因.

# idgo服务正确性和高可用保障措施

## 1. 压力测试结果
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

## 2.ID生成正确性验证

- 模拟4个进程（cmd/example.go）以每秒生成100个ID的频率并发向MySQL IDGEN服务申请ID，并将生成的ID写入MySQL。
测试16小时后，生成330万个ID，未发现有重复ID。

- idgo已经部署在线上稳定运行。

## 3.ID生成服务宕机后的恢复方案

当idgo服务意外宕机后，可以切从库，然后将idgo对应的key加上适当的偏移量。

## 4. 使用方法

参考cmd/example.go文件使用, 用起来很简单. :)

编译并运行example.go
```
. ./dev.sh
make
./bin/cmd

```

# License

MIT

