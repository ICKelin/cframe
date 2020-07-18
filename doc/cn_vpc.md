# 腾讯云，阿里云VPC互联操作
使用cframe解决多云场景下VPC互联，在本次测试当中，通过cframe打通阿里云vpc和腾讯云VPC。

本次测试包含两个阿里云VPC和一个腾讯云VPC。具体网络拓扑参考如下

![images/cframe1.1.0.jpg](images/cframe1.1.0.jpg)

需要注意三个VPC网段，避免网段冲突。

## 云服务器信息

- VPC1 阿里云深圳

vpc 网段: 172.18.0.0/16

| 内网ip | 角色 |
|:--| :-- |:--|
| 172.18.171.245 | edage + controller|
| 172.18.171.247 | - |

2. VPC2 阿里云香港

vpc 网段: 172.31.0.0/16

| 内网ip | 角色 |
|:--|:--|
| 172.31.185.158 | edage |
| 172.31.185.159 | - |

3. VPC3 腾讯云广州

vpc cidr: 172.20.0.0/16

| 内网ip | 角色 |
|:--|:--|
| 172.20.0.9  | edage |
| 172.20.0.13 | - |

## 第一步： 构建项目
```
./build.sh
```

生成的二进制文件以及配置文件均在dist目录下。

## 第二步: 应用程序配置

1. controller需要在配置文件当中配置三个VPC名称以及CIDR信息。

```

listen_addr=":58422"
api_addr=":12345"

etcd = [
    "127.0.0.1:2379"
]


[[edages]]
name="sz-1"
# VPC1 public ip
host_addr = "$PIP1:58423"
cidr = "172.18.0.0/16"

[[edages]]
name="hk-1"
# VPC2 public ip
host_addr = "$PIP2:58423"
cidr = "172.31.0.0/16"

[[edages]]
name="gz-3"
host_addr = "$PIP3:58423"
cidr = "172.20.0.0/16"
```

出于安全目的，涉及到公网IP的部分均采用变量代替，使用时记得进行替换。

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
参考上节 **云服务器信息** 当中的角色字段将controller和edage运行

### VPC路由配置
在云服务厂商VPC配置下配置路由信息，需要将另外两个VPC的网段都加入到路由当中，下一跳指向所在VPC的edage节点实例。

具体参照云服务厂商文档:

- [阿里云自定义路由](https://help.aliyun.com/document_detail/87057.html?spm=a2c4g.11186623.6.585.40b01db2U1KwfP)
- [腾讯云管理路由表](https://cloud.tencent.com/document/product/215/36682)


### 测试
一切就绪之后，就可以对网络连通性测试，cframe的目的是让三个网络互通，所以在任意一个vpc的任意一台云服务器上ping对端两个vpc的任何一个ip，都可以ping通。