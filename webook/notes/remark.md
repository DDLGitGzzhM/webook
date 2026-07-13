
middleware 所有业务都需要关心的东西，就是 AOP 的解决方案

todo 设计并实现一个 GIN 的插件库
1. Web治理 > 熔断限流降级
2. 可观测性 > 日志 metric tracing


Gin 面试题

1. 什么是 Gin 的 middleware ？用来解决什么问题
   - 支持 IP 限流
   - 支持 VIP 降级实现
2. 什么是跨域问题，怎么解决 ？
3. 跨域问题需要设置哪些头部 ？ 


ORM
ORM的全称是对象关系映射（Object-Relational Mapping）。
简单来说，它就像一座桥梁。它帮你把程序里的数据对象（代码）自动转换成关系型数据库（表格）里的数据

引入 Service-Repository-Dao 三层架构 其中 service, repository 参考的是 DDD 设计

service 层：领域服务
repository 层：数据对象的存储
dao : 数据库的操作
domain 层：领域对象

面试题
1. 什么是 Cookie，什么事 Session
2. Cookie 和 Session 比起来有什么缺点
3. Session Id可以放在哪 ？这个问题 你要记得提起Cookie 禁用的问题
4. 用户密码加密算法有什么注意事项？你用的什么
5. 怎么做登陆校验？核心是 GIN 的middleware