

## 点赞/阅读/收藏

如果我有一个点赞/浏览/收藏表，其中 (biz,bizid,Cnt)这三个关键字段， 对于 建立联合索引应该是 (biz,bizid)还是(bizid,biz) ，另外为什么如果我想要获取cnt前10的数据，为什么没必要在 Cnt上面建立索引


增加计数

下面这段代码 有并发问题，如果两个 goroutine 增加了同一个bizid的数据，那么就会有问题

如果 cnt = 10, 两个进来之后只会更新成 11 ，但是我们想要的是 12

```go
err : db.Where("biz = ? and bizid = ?", biz, bizid).First(&like).Error
if err != nil {}
cnt := like.Cnt + 1 
db.Where("biz = ? and bizid = ?", biz, bizid).Update("Cnt", cnt)
```

解决办法

1. for update 最好不要
2. SQL 自动支持 update = a = a + 1

```go
dao.db.
	Where("biz = ? and bizid = ?", biz, bizid).
	Updates(map[string]any{
	"Cnt": gorm.Expr("Cnt + ?", 1),
	"UpdatedAt": time.Now(),
})
```

那么如果没有这一行呢,所以这里应该是一个 `upsert` 的语意

```go
db.Clauses(clause.OnConflict{
	DoUpdates: clause.Assignments(map[string]any{
"Cnt": gorm.Expr("Cnt + ?", 1),
"UpdatedAt": time.Now(),
}))
}).Create()
```

业务耦合的情况下的区别

1. IO
2. 磁盘
3. 写友好 。读友好

```go
(biz,bizid,Cnt,cntTYpe)

(biz,bizid,readCnt,likeCnt,CollectCnt)
```

怎么理解 QPS

1. B站如果有100w的用户，那么他们访问同一个 rest ， 是不是这个rest就有100w个qps，这里面是顺序的还是并发的 ？ 
2. 如果只有一个用户访问一个rest，但是开启了 100w个gourtione，这里面应该是并发的？应该和100w个人一样 ？

缓存方案

方案1 :

1. map[string]int . hgetall 

方案 2 :

1. key1_read_cnt 
2. key1_collect_cnt 
3. key1_like_cnt

这两个的区别就是 一次查询和3次查询，但是方案2可以使用lua进行处理 ， 方案2一个查询不也只是对应一个链接吗？在网络那边有点差异？