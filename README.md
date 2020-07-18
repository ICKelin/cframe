## cframe
cframe is a vpc peering product which is used connect different cloud provider(eg: aws, aliyun, tencent cloud).

## How it works
cframe is a CUPS like project. it contains `controller` and `edage`.

edage is our data plane, it gets other edages from controller, routes, and finally forwards packet to peer via UDP.

controller is our control plane, she export http api to users to configure edage and dispatch configuration to edage via tcp longlive connection.

![doc/images/cframe1.0.0](doc/images/cframe1.1.0.jpg)

## Get start
In this case, will will use cframe to connect aliyun and tencent cloud, we make 3 vpc connects via internal ip

### VPCs and cloud server
- VPC1, aliyun ShenZhen.

vpc cidr: 172.18.0.0/16

| internal ip | role |
|:--| :-- |
| 172.18.171.245 | edage + controller|
| 172.18.171.247 | - |

2. VPC2, aliyun HongKong

vpc cidr: 172.31.0.0/16

| internal ip| role |
|:--|:--|
| 172.31.185.158 | edage |
| 172.31.185.159 | - |

3. VPC3 tengcent cloud GuangZhou

vpc cidr: 172.20.0.0/16

| internal ip | role |
|:--|:--|
| 172.20.0.9  | edage |
| 172.20.0.13 | - |

our goal is to connect other two vpc via internal ip address. 

### step1: build controller and edage

```
./build.sh
```

it will created dist folder, build, copy configuration file.

```
dist
├── controller
├── controller.toml
├── edage
└── edage.toml
```

### step2: deploy controller in cloud server

controller.toml:

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

replace `$PIP1`, `$PIP3`, `$PIP3` to your vpc public ip.

### step3: deploy edage

VPC1 edage.toml

```
controller="$CONTROLLER_PIP:58422"
name = "sz-1"
listen_addr=":58423"

```

VPC2 edage.toml

```
controller="$CONTROLLER_PIP:58422"
name = "hk-1"
listen_addr=":58423"
```

VPC3 edage.toml

```
ubuntu@VM-0-9-ubuntu:~$ cat edage/config.toml 
controller="$CONTROLLER_PIP:58422"
name = "gz-3"
listen_addr=":58423"
```

replace `$CONTROLLER_PIP` with your controller public ip.

### step4: add vpc route entry

we need to add peers cidr to VPC route to make sure the traffic transfer to our edage node

- in VPC1, add VPC2, VPC3 cidr to VPC route, nexthop to VPC1.edage instance
- in VPC2, add VPC1, VPC3 cidr to VPC route, nexthop to VPC2.edage instance
- in VPC3, add VPC1, VPC2 cidr to VPC route, nexthop to VPC3.edage instance

### step5: test network connection
Here are some testcases.

- in VPC1, ping VPC2

```
ping 172.31.185.158
ping 172.31.185.159
```

- in VPC1, ping VPC3

```
ping 172.20.0.9
ping 172.20.0.13
```

- in VPC2, ping VPC1

```
ping 172.18.171.245
ping 172.18.171.247
```

- in VPC2, ping VPC3

```
ping 172.20.0.9
ping 172.20.0.13
```

- in VPC3, ping VPC1

```
ping 172.18.171.245
ping 172.18.171.247
```

- in VPC3, ping VPC2

```
ping 172.31.185.158
ping 172.31.185.159
```
