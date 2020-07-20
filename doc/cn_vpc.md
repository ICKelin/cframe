# 腾讯云，阿里云VPC互联操作
使用cframe解决多云场景下VPC互联，在本次测试当中，通过cframe打通阿里云vpc和腾讯云VPC。

本次测试包含两个阿里云VPC和一个腾讯云VPC。具体网络拓扑参考如下

![images/cframe1.1.0.jpg](images/cframe1.1.0.jpg)

需要注意三个VPC网段，避免网段冲突。

## 云服务器信息

- VPC1 阿里云深圳

vpc 网段: 172.18.0.0/16

| 内网ip | 角色 |
|:--| :-- |
| 172.18.171.245 | edge + controller|
| 172.18.171.247 | - |

2. VPC2 阿里云香港

vpc 网段: 172.31.0.0/16

| 内网ip | 角色 |
|:--|:--|
| 172.31.185.158 | edge |
| 172.31.185.159 | - |

3. VPC3 腾讯云广州

vpc cidr: 172.20.0.0/16

| 内网ip | 角色 |
|:--|:--|
| 172.20.0.9  | edge |
| 172.20.0.13 | - |

## 第一步： 构建项目
```
./build.sh
```

生成的二进制文件以及配置文件均在dist目录下。

## 第二步: 应用程序配置

1. controller需要在配置文件当中配置三个VPC名称~~以及CIDR信息~~。

**最新版本不支持在配置文件配置cidr信息，仅能通过api操作cidr**

```

listen_addr=":58422"
api_addr=":12345"

etcd = [
    "127.0.0.1:2379"
]

```

controller需要指定etcd的endpoints地址，apiserver监听地址，与edge建立连接的tcp地址

2. aliyun-sz 节点配置

```
controller="$CONTROLLER_PIP:58422"
name = "sz-1"
listen_addr=":58423"

```

3. aliyun-hk 节点配置

```
controller="$CONTROLLER_PIP:58422"
name = "hk-1"
listen_addr=":58423"

```

4. tencent-gz 节点配置

```
controller="$CONTROLLER_PIP:58422"
name = "gz-1"
listen_addr=":58423"

```

### 运行应用
参考上节 **云服务器信息** 当中的角色字段将controller和edge运行

```
./controller -c config.toml

./edge -c config.toml
```


程序启动之后，通过以下三个curl命令新增三个VPC信息以及CIDR信息。

```
curl "http://127.0.0.1:12345/api-service/v1/edge/add" -X "POST" -d '{"name": "sz-1", "hostaddr": "$PIP1:58423", "cidr": "172.18.0.0/16"}' -H "Content-Type: application/json" 

curl "http://127.0.0.1:12345/api-service/v1/edge/add" -X "POST" -d '{"name": "hk-1", "hostaddr": "$PIP2:58423", "cidr": "172.31.0.0/16"}' -H "Content-Type: application/json" 

curl "http://127.0.0.1:12345/api-service/v1/edge/add" -X "POST" -d '{"name": "gz-3", "hostaddr": "$PIP3:58423", "cidr": "172.20.0.0/16"}' -H "Content-Type: application/json"

```

执行完成之后，通过api查询结果。

```
1. 查询edge信息
curl "http://127.0.0.1:12345/api-service/v1/edge/list"|jq
[
  {
    "name": "sz-1",
    "comment": "",
    "cidr": "172.18.0.0/16",
    "host_addr": "$PIP1:58423",
    "status": 0
  },
  {
    "name": "gz-3",
    "comment": "",
    "cidr": "172.20.0.0/16",
    "host_addr": "$PIP2:58423",
    "status": 0
  },
  {
    "name": "hk-1",
    "comment": "",
    "cidr": "172.31.0.0/16",
    "host_addr": "$PIP3:58423",
    "status": 0
  }
]
```

**出于安全目的，涉及到公网IP的部分均采用变量代替，使用时记得进行替换。**


### VPC路由配置
在云服务厂商VPC配置下配置路由信息，需要将另外两个VPC的网段都加入到路由当中，下一跳指向所在VPC的edge节点实例。

具体参照云服务厂商文档:

- [阿里云自定义路由](https://help.aliyun.com/document_detail/87057.html?spm=a2c4g.11186623.6.585.40b01db2U1KwfP)
- [腾讯云管理路由表](https://cloud.tencent.com/document/product/215/36682)


### 测试
一切就绪之后，就可以对网络连通性测试，cframe的目的是让三个网络互通，所以在任意一个vpc的任意一台云服务器上ping对端两个vpc的任何一个ip，都可以ping通。

```
ubuntu@VM-0-13-ubuntu:~$ ifconfig eth0 --------------------------------> 在172.20.0.13(VPC3)上执行
eth0      Link encap:Ethernet  HWaddr 52:54:00:d7:57:54
          inet addr:172.20.0.13  Bcast:172.20.255.255  Mask:255.255.0.0
          inet6 addr: fe80::5054:ff:fed7:5754/64 Scope:Link
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:655502 errors:0 dropped:0 overruns:0 frame:0
          TX packets:677087 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:74802592 (74.8 MB)  TX bytes:92601091 (92.6 MB)

ubuntu@VM-0-13-ubuntu:~$ ping 172.18.171.247  ---------------------> ping VPC1
PING 172.18.171.247 (172.18.171.247) 56(84) bytes of data.
64 bytes from 172.18.171.247: icmp_seq=1 ttl=62 time=8.93 ms
^C
--- 172.18.171.247 ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 8.937/8.937/8.937/0.000 ms
ubuntu@VM-0-13-ubuntu:~$ ping 172.18.171.245  ---------------------> ping VPC1
PING 172.18.171.245 (172.18.171.245) 56(84) bytes of data.
64 bytes from 172.18.171.245: icmp_seq=1 ttl=63 time=8.56 ms
64 bytes from 172.18.171.245: icmp_seq=2 ttl=63 time=8.48 ms
^C
--- 172.18.171.245 ping statistics ---
2 packets transmitted, 2 received, 0% packet loss, time 1001ms
rtt min/avg/max/mdev = 8.483/8.523/8.564/0.100 ms
ubuntu@VM-0-13-ubuntu:~$ ping 172.31.185.159  ---------------------> ping VPC2
PING 172.31.185.159 (172.31.185.159) 56(84) bytes of data.
64 bytes from 172.31.185.159: icmp_seq=1 ttl=62 time=67.2 ms
64 bytes from 172.31.185.159: icmp_seq=2 ttl=62 time=66.6 ms
^C
--- 172.31.185.159 ping statistics ---
2 packets transmitted, 2 received, 0% packet loss, time 1001ms
rtt min/avg/max/mdev = 66.611/66.926/67.242/0.408 ms
ubuntu@VM-0-13-ubuntu:~$ ping 172.31.185.158  ---------------------> ping VPC2
PING 172.31.185.158 (172.31.185.158) 56(84) bytes of data.
64 bytes from 172.31.185.158: icmp_seq=1 ttl=63 time=66.7 ms
64 bytes from 172.31.185.158: icmp_seq=2 ttl=63 time=66.7 ms
^C
--- 172.31.185.158 ping statistics ---
2 packets transmitted, 2 received, 0% packet loss, time 1001ms
rtt min/avg/max/mdev = 66.706/66.712/66.719/0.258 ms
ubuntu@VM-0-13-ubuntu:~$
```
