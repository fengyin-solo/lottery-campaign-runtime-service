# Bug

通知接口携带 nil 动态指针时被视为可用提供方，调用发生 panic，恢复后回执与错误互相矛盾。

# 触发

先用 typed-nil 通知提供方发送消息，再换成有效提供方发送下一条消息。

# 错误信息

`notification provider panic: nil notifier dereference`
