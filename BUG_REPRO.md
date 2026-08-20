# Bug

最后一份奖品的可变快照跨并发请求共享，余量检查与扣减不在同一同步边界内。

# 触发

让两个领取请求先读取余量为一的相同奖品，在同步信号后同时执行预留。

# 错误信息

`snapshot clone shares state: original Remaining:0 copy Remaining:0`
