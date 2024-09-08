## 前言
一直以来都来研究`redis`的设计理念和实现思路，随着时间的推移有了想自己实现的想法，由于笔者本身是一名`java`开发，对于`C语言`中的某些理念不是非常熟悉，于是折中选用`go语言`进行实现，而这份文档将会记录笔者实现`mini-redis`的一些开发思路和实现的介绍。
这个项目笔者无论从函数名还是整体思路都基本沿用了原生`redis`的规范，并且为了让读者更加直观的了解`redis`的实现核心脉络，笔者也在实现时也进行了一定简化，希望对那些想了解redis但是又不太熟悉`C语言`的开发朋友有所帮助。


## 如何部署运行mini-redis

这里笔者以`Linux`为例，我们找到项目中的`build-linux.sh`键入如下指令:

```bash
./build-linux.sh
```

执行该指令之后，我们的项目会生成一个bin的文件夹，其内部一个`redis-server`的可执行文件，此时我们通过Linux服务器执行`./redis-server`，最终项目会输出如下内容，说明项目成功运行了：

```bash
sharkchili@DESKTOPJ:~$ ./redis-server
2024/09/08 23:07:09 init redis server
2024/09/08 23:07:09 load redis server config
2024/09/08 23:07:09 this redis server address: localhost:6379
2024/09/08 23:07:09 event loop is listening and waiting for client connection.

```


此时我们就可以通过redis-cli连接mini-redis了：

```bash
sharkchili@DESKTOP-xxxx:~/redis/src$ ./redis-cli
127.0.0.1:6379> command
1) "COMMAND"
2) "PING"
127.0.0.1:6379> ping
PONG
127.0.0.1:6379>

```




## 目前的开发进度

截至目前，笔者完成了下面几个核心模块的开发：

- `mini-redis`服务端ip端口绑定初始化。
- `redis-cli`连接时将其封装为`redisClient`，并能够接收其命令请求。
- 解析`redis-cli`通过`RESP`协议发送的命令请求。
- 支持客户端键入command指令并回复当前服务端支持的指令集。
- 支持客户端键入PING指令感知mini-redis是否可用。


后续开发计划:

+ [ ] 字典、列表、有序集合、字符串等数据结构
+ [ ] 字符串常规`GET`和`SET`指令
+ [ ] 列表有所操作指令
+ [ ] 字典所有操作指令
+ [ ] 有序集合所有操作指令
+ [ ] `AOF`持久化和重载机制
+ [ ] `LRU`缓存置换算法


## 如何阅读源码

该项目笔者还在不断和开发和迭代中，在开发过程中的设计和实现也都会不断输出为文档，读者可以按需取用下面的文章来了解笔者的开发过程和实现思路：


来聊聊我用go手写redis这件事:<https://mp.weixin.qq.com/s?__biz=MzkwODYyNTM2MQ==&mid=2247486169&idx=1&sn=9b562eca113fbbe02d07f9ae2ccf0e79&chksm=c0c65e67f7b1d77166b7ff9a7a0403fb02d39ae1a111a73e12f7f1cdc3e79e91b2d116a5908a#rd>


## 关于我

Hi，我是 **sharkChili** ，是个不断在硬核技术上作死的技术人，是 **CSDN的博客专家** ，也是开源项目 **Java Guide** 的维护者之一，熟悉 **Java** 也会一点 **Go** ，偶尔也会在 **C源码** 边缘徘徊。写过很多有意思的技术博客，也还在研究并输出技术的路上，希望我的文章对你有帮助，非常欢迎你关注我的公众号： **写代码的SharkChili** 。

因为近期收到很多读者的私信，所以也专门创建了一个交流群，感兴趣的读者可以通过上方的公众号获取笔者的联系方式完成好友添加，点击备注  **“加群”**  即可和笔者和笔者的朋友们进行深入交流。


![image](https://github.com/user-attachments/assets/ed27dbdf-f0da-40b3-bb4e-cf948c4611a3)
