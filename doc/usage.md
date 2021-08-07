## cframe使用
在本文档当中，会通过使用cframe来将位于深圳阿里云的VPC和位于香港的AWS VPC打通来具体说明如何cframe的使用步骤。

首先网络拓扑如下:

![](images/demo.jpg)

其中阿里云深圳的VPC网段为`172.18.0.0/16`，AWS香港的VPC网段为`172.30.0.0/16`，最终的目的是能实现这两个网段相互之间能够访问。

## 安装controller
cframe需要依赖etcd，所以首先需要安装etcd，etcd安装完成之后，下载controller并运行。

controller的配置文件:

```yaml
listen_addr = ":58422"
etcd = [
  # replace with etcd endpoints
  "172.18.171.247:2379"
]

[log]
level = "info"
path = "controller.log"
days = 5
```

大部分情况下，您只需要修改etcd的endpoints信息即可将controller运行成功。关于controller，您需要注意的地方有：

- 需要运行在可以通过公网IP和端口访问到的机器当中
- 请注意检查安全组是否开通58422/tcp端口

配置文件生成之后，只需要
`./controller -c config.toml` 运行controller即可。

除了直接命令使用之外，您也可以使用docker来进行运行，使用docker来运行的一个好处是controller由于异常崩溃时可以自动重启。步骤大同小异，这里不在赘述。

## 使用cfctl配置基本信息
在运行程序之前，首先需要使用`cfctl`工具创建好基本信息，包括：

1. namespace，namespace的目的是为了与其他配置隔离，关于namespace的细节可以参考[namespace隔离设计]()

2. 将两端edge所在的云服务器监听的公网IP和端口进行登记

接下来先进行namespace的创建， 

```sh
➜  ~ cfctl namespace
NAME:
   cfctl namespace - manage namespace

USAGE:
   cfctl namespace command [command options] [arguments...]

COMMANDS:
   add      add a new namespace
   del      delete a namespace
   list     list all namespaces
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)

# 创建名为demons的namespace
➜  ~ cfctl namespace add --name demons
create namespace demons, secret TkeqZ+ZCQd2gQwEJjbA8Sg== OK

# 创建成功之后，会返回secret，整个使用过程中都会涉及到secret

```

创建完namespace之后，需要登记edge节点的信息，此处包含深圳阿里云和香港aws两个edge。

```sh
➜  ~ cfctl edge add
NAME:
   cfctl edge add - add a new edge

USAGE:
   cfctl edge add [command options] [arguments...]

OPTIONS:
   --namespace value, --ns value  (default: "default")
   --name value
   --listener value               edge listener, eg: 1.2.3.4:58423
   --cidr value                   eg: 172.18.0.0/16
   --help, -h                     show help (default: false)

```

cfctl的edge新增需要四个参数:

- namespace - 也就是上一步当中创建的namespace名称，此处为demons
- name - edge节点的名称，edge的唯一标识
- listener - edge节点监听的公网ip和udp端口
- cidr - edge节点所在的网段

了解了这四个参数之后，下面先创建深圳阿里云VPC的edge信息

```sh
➜  ~ cfctl edge add --namespace=demons --name=edge-aliyun-sz --listener=47.115.82.137:38424 --cidr=172.18.0.0/16
create edge 47.115.82.137:38424 cidr 172.18.0.0/16 OK

➜  ~ cfctl edge list --namespace=demons
edge list:
      Name            Listener                  CIDR
-----------------------------------------------------------
1     edge-aliyun-sz  47.115.82.137:38424       172.18.0.0/16
```

接下来按照同样的方式，创建香港AWS VPC的edge信息

```sh
➜  ~ cfctl edge add --namespace=demons --name=edge-aws-hk --listener=18.163.79.238:38423 --cidr=172.30.0.0/16
create edge 18.163.79.238:38423 cidr 172.30.0.0/16 OK
➜  ~ cfctl edge list --ns demons
edge list:
      Name            Listener                  CIDR
-----------------------------------------------------------
1     edge-aliyun-sz  47.115.82.137:38424       172.18.0.0/16
2     edge-aws-hk     18.163.79.238:38423       172.30.0.0/16
```

## 运行edge节点

万事具备，只差把edge节点拉起来了，edge节点没有配置文件，需要的几个参数都是通过环境变量的方式传入。

- listen - 本地监听的udp地址，需要与之前步骤当中创建的edge信息里面的listener端口对应，此处为:38424和:38423
- controller - controller的监听地址
- secret - namespace的secret
- namespace - namespace的名称
- name - edge节点名称

那么接下来还是先从深圳阿里云开始，将edge节点拉起来。

```sh
namespace=demons secret=TkeqZ+ZCQd2gQwEJjbA8Sg== name=edge-aliyun-sz controller=demo.notr.tech:58422 listen=:38424 nohup ./edge &
```

执行成功之后，系统会多出一条发往aws香港VPC的路由

```sh
route -n |grep cframe
172.30.0.0      0.0.0.0         255.255.0.0     U     0      0        0 cframe.0
```

使用同样的方式运行aws香港的edge程序。

`namespace=demons secret=TkeqZ+ZCQd2gQwEJjbA8Sg== name=edge-aws-hk controller=demo.notr.tech:58422 listen=:38423 nohup ./edge &`

运行成功之后，同样会新增一条路由。

```sh
[root@ip-172-30-102-132 ec2-user]# route -n |grep cframe
172.18.0.0      0.0.0.0         255.255.0.0     U     0      0        0 cframe.0
```

至此，所有操作都已经执行完成了。

## 测试验证

- 在深圳阿里云ping香港aws的内网ip
```
root@iZwz97kfjnf78copv1ae65Z:~# ping 172.30.102.132
PING 172.30.102.132 (172.30.102.132) 56(84) bytes of data.
64 bytes from 172.30.102.132: icmp_seq=1 ttl=255 time=130 ms
64 bytes from 172.30.102.132: icmp_seq=2 ttl=255 time=135 ms
64 bytes from 172.30.102.132: icmp_seq=3 ttl=255 time=133 ms
64 bytes from 172.30.102.132: icmp_seq=4 ttl=255 time=129 ms
64 bytes from 172.30.102.132: icmp_seq=5 ttl=255 time=135 ms
^C
--- 172.30.102.132 ping statistics ---
5 packets transmitted, 5 received, 0% packet loss, time 4002ms
rtt min/avg/max/mdev = 129.712/133.009/135.761/2.456 ms
```

- 在香港aws ping深圳阿里云的一台云服务器的内网IP

```sh
[ec2-user@ip-172-30-102-132 ~]$ ping 172.18.171.247
PING 172.18.171.247 (172.18.171.247) 56(84) bytes of data.
64 bytes from 172.18.171.247: icmp_seq=1 ttl=64 time=131 ms
64 bytes from 172.18.171.247: icmp_seq=2 ttl=64 time=133 ms
64 bytes from 172.18.171.247: icmp_seq=3 ttl=64 time=135 ms
^C
--- 172.18.171.247 ping statistics ---
4 packets transmitted, 3 received, 25% packet loss, time 3001ms
rtt min/avg/max/mdev = 131.079/133.529/135.791/1.973 ms
```

通过这两轮测试验证，网络连通性是正常的。

## 总结
本文档中详细介绍了如何部署cframe的controller和edge组件，并使用cfctl管理edge和namespace信息，在此对整体流程做一个简要梳理。

- 首先处于租户隔离的目的，需要创建namespace
- 创建完namespace之后，需要将对应的edge信息登记，以便controller识别哪些是合法的edge连接
- 运行controller
- 运行edge节点，需要用到namespace的名称和secret，edge的名称以及监听地址等信息。

希望这篇文章能够帮助您成功使用cframe，如果没有，请联系作者或者提交issue。