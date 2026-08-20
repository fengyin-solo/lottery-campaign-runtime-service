# Bug

池化审计请求没有清除追踪身份，无追踪号的下一请求会继承上一位用户的信息。

# 触发

先审计带 `trace-secret` 的用户 A，再审计不带追踪号的用户 B，并读取两条记录。

# 错误信息

`second audit leaked identity: {user-b campaign-b trace-secret}`
