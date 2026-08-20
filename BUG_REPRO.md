# Bug

带说明文字包装后的临时兑换错误被当作永久错误，重试提前结束并产生错误提交状态。

# 触发

让第一次兑换返回包装后的 `temporary redemption failure`，第二次调用预期成功。

# 错误信息

`redeem returned gateway: temporary redemption failure`
