# cframe
[English Readme](EN_README.md)

cframe是一个用于连接VPC的网络基础项目，使用cframe可以将同一云服务厂商，不同云服务厂商的VPC之间进行网络互通，VPC和VPC之间可以通过内网地址进行互联互通，保证数据安全，除此之外。

理论上除了VPC互联之外，也可以将VPC和IDC机房网络打通，设置是将多个办公室网络打通。

![doc/images/cframe1.0.0](doc/images/cframe1.1.0.jpg)

## cframe原理
cframe包括两个角色：

- controller
- edage

controller也即是控制平面，维护全局路由表以及路由表下发。edage也称为边缘节点，也即是常说的数据平面。只负责根据controller的路由信息进行转发，选择正确的对端节点(peer)进行转发，转发协议目前采用的是UDP，后续可能考虑使用QUIC或者KCP这类ARQ协议。

## 快速开始
此处使用cframe连接三个VPC作为例子展示cframe的作用。

详细参考[腾讯云，阿里云VPC互联操作](doc/cn_vpc.md)

## Bugs反馈
有任何bug或者建议，可以给我们提issue或者pr。

## 合作
如果在使用过程中有任何问题，可以通过以下方式和我取得联系

- Email: 995139094@qq.com
- Wechat: zyj995139094
