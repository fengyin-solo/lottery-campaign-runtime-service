# Bug

异步受众批次保留调用方切片的底层数组，调用方复用数组后会改写已提交名单。

# 触发

提交包含 `u-1`、`u-2` 的批次，等待消费者阻塞后复用原数组写入下一批，再释放消费者。

# 错误信息

`submitted audience changed after caller reuse: [u-next u-2]`
