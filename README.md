[![Build Status](https://travis-ci.org/ICKelin/cframe.svg?branch=master)](https://travis-ci.org/ICKelin/cframe) ![goreport](https://goreportcard.com/badge/github.com/ICKelin/cframe)

# cframe
[English Readme](EN_README.md)

cframe是一个用于连接公有云VPC，数据中心，企业分支的网络基础项目，通过cframe达到多点网络互联互通的目的，同时，为了解决海外公有云VPC/数据中心节点传输速度慢的的问题，cframe内部舍弃tcp传输，暂时使用kcp协议来进行。

![doc/images/cframe1.0.0](doc/images/cframe1.1.0.jpg)

## cframe原理
![arch](doc/images/arch.jpg)

cframe包括两个角色：

- controller
- edge

controller也即是控制平面，维护全局路由表以及路由表下发。edge也称为边缘节点，也即是常说的数据平面。只负责根据controller的路由信息进行转发，选择正确的对端节点(peer)进行转发，转发协议目前采用的是UDP，后续可能考虑使用QUIC或者KCP这类ARQ协议。


## 快速开始
此处使用cframe连接三个VPC作为例子展示cframe的作用。

详细参考[腾讯云，阿里云VPC互联操作](doc/cn_vpc.md)

## Bugs反馈
有任何bug或者建议，可以给我们提issue或者pr。

## 合作
如果在使用过程中有任何问题，可以通过以下方式和我取得联系

- Email: 995139094@qq.com
- Wechat: zyj995139094
