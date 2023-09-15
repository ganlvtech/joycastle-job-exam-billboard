# JoyCastle 后端程序笔试题

* 你开发了一个游戏，日活跃用户在 10 万人以上。请设计一个活动排行榜系统。
* 在每月活动中，玩家得到的活动总分为 0 到 10000 之间的整数。
* 在每月活动结束之后，需要依据这一活动总分，从高到低为玩家建立排行榜。
* 如果多位玩家分数相同，则按得到指定分数顺序排序，先得到的玩家排在前面。
* 系统提供玩家名次查询接口，玩家能够查询自己名次前后 10 位玩家的分数和名次。
* 请使用 UML 图或线框图表达设计，关键算法可使用流程图或伪代码表达。

## 分析

日活 10 万，月活按 100 万计算。玩家数据有 UID、活动分数、获得分数时的时间戳、排名这几个数据，直接读取到内存中大概会占用 32MB 内存，可以接受。

这个需求里一个关键的点在于，每月活动结束之后才建立排行榜，因此实现起来就很简单了。如果是实时动态更新的排行榜，会稍微复杂一些。

使用一个数组记录上面的几项数据，按分数倒序、时间正序排序，然后使用 Hash 表记录 UID -> 数组下标的映射关系。

查询时，先根据玩家 UID 查到他的数组下标，然后再查出下标 - 10 ~ 下标 + 10 的数据，返回给客户端。

说明：数据源可以使用 MySQL、Redis 或者 CSV，我这里使用 CSV 因为他最简单并且依赖最少。

说明：我没有使用其他 HTTP 框架，而使用 Go 内置的 HTTP 框架了，因为他最简单并且依赖最少，可以直接运行。如果项目比较复杂，路径比较的话，可以考虑使用 Gin 等框架。

## 运行

先运行

```bash
go run ./cmd/generate_rand_data/
```

生成用于测试的数据 `data.csv`（大约 23MB）

然后再运行

```bash
go run main.go
```

启动 HTTP 服务器。

启动后大概占用 80MB 内存。

## 使用演示

访问 http://127.0.0.1:8000/nearby_ranks?uid=123

```json
{
  "Code": 0,
  "Msg": "OK",
  "Data": [
    {
      "UID": 176760,
      "Score": 1631,
      "Timestamp": 1694787621,
      "Rank": 836837
    },
    // ......
    {
      "UID": 123,
      "Score": 1631,
      "Timestamp": 1694994996,
      "Rank": 836847
    },
    // ......
    {
      "UID": 369464,
      "Score": 1631,
      "Timestamp": 1695291930,
      "Rank": 836857
    }
  ]
}
```

## 性能

因为排行榜是固定的，这个服务器完全可以水平拓展。另外日活 10 万，峰值在线按 5 万算，Golang 的性能单机肯定能抗住。

这是我在 16 核 32 线程的个人电脑上的压测

```plain
> hey -z 10s 'http://127.0.0.1:8000/nearby_ranks?uid=123'

Summary:
  Total:        10.0003 secs
  Slowest:      0.0180 secs
  Fastest:      0.0001 secs
  Average:      0.0005 secs
  Requests/sec: 131283.8672
  
  Total data:   1823590320 bytes
  Size/request: 1823 bytes

Response time histogram:
  0.000 [1]     |
  0.002 [997724]        |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.004 [2026]  |
  0.005 [156]   |
  0.007 [31]    |
  0.009 [16]    |
  0.011 [27]    |
  0.013 [14]    |
  0.014 [1]     |
  0.016 [2]     |
  0.018 [2]     |


Latency distribution:
  10% in 0.0002 secs
  25% in 0.0003 secs
  50% in 0.0004 secs
  75% in 0.0005 secs
  90% in 0.0006 secs
  95% in 0.0007 secs
  99% in 0.0010 secs

Details (average, fastest, slowest):
  DNS+dialup:   0.0000 secs, 0.0001 secs, 0.0180 secs
  DNS-lookup:   0.0000 secs, 0.0000 secs, 0.0000 secs
  req write:    0.0000 secs, 0.0000 secs, 0.0103 secs
  resp wait:    0.0004 secs, 0.0001 secs, 0.0179 secs
  resp read:    0.0000 secs, 0.0000 secs, 0.0141 secs

Status code distribution:
  [200] 1000000 responses
```
