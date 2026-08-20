# 抽奖活动服务 (lottery)

纯 Go 标准库实现的抽奖活动后端服务，零第三方依赖。

## 运行

```bash
cd origin
go run ./cmd/server
```

默认监听 `:8080`，可通过环境变量 `PORT` 或 `ADDR` 修改。

## 实体

- Activity 活动
- Prize 奖品
- Entry 参与记录
- Winner 中奖记录
- User 用户

## 运行流模块

`internal/runtimeflow` 提供抽奖后台任务的并发与生命周期示例，包括活动导出、受众批处理、奖品领取、通知、重试、审计和库存预留。模块按 `model`、`store`、`worker`、`service` 分层，测试覆盖取消传播、异步数据所有权、资源释放和事务可见性。

## API 列表

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/activities | 创建活动 |
| GET | /api/activities | 活动列表（支持 status、keyword 筛选与分页） |
| GET | /api/activities/{id} | 活动详情 |
| PUT | /api/activities/{id} | 更新活动 |
| DELETE | /api/activities/{id} | 删除活动 |
| POST | /api/activities/{id}/transition | 活动状态流转（draft→active→ended） |
| POST | /api/prizes | 创建奖品 |
| GET | /api/prizes | 奖品列表（支持 activity_id、keyword 筛选与分页） |
| GET | /api/prizes/{id} | 奖品详情 |
| PUT | /api/prizes/{id} | 更新奖品 |
| DELETE | /api/prizes/{id} | 删除奖品 |
| POST | /api/entries | 创建参与记录 |
| GET | /api/entries | 参与记录列表（支持 activity_id、user_id、result 筛选与分页） |
| GET | /api/entries/{id} | 参与记录详情 |
| PUT | /api/entries/{id} | 更新参与记录 |
| DELETE | /api/entries/{id} | 删除参与记录 |
| POST | /api/winners | 创建中奖记录 |
| GET | /api/winners | 中奖记录列表（支持 activity_id、user_id、status 筛选与分页） |
| GET | /api/winners/{id} | 中奖记录详情 |
| PUT | /api/winners/{id} | 更新中奖记录 |
| DELETE | /api/winners/{id} | 删除中奖记录 |
| POST | /api/winners/{id}/claim | 领取奖品（won→claimed） |
| POST | /api/winners/batch-expire | 批量将未领取中奖记录设为过期 |
| POST | /api/users | 创建用户 |
| GET | /api/users | 用户列表（支持 keyword 筛选与分页） |
| GET | /api/users/{id} | 用户详情 |
| PUT | /api/users/{id} | 更新用户 |
| DELETE | /api/users/{id} | 删除用户 |
| POST | /api/lottery/draw | 单次抽奖 |
| POST | /api/lottery/batch-draw | 批量抽奖 |
| GET | /api/stats/activities/{id} | 活动统计（参与人数、中奖人数、奖品分布） |
| GET | /api/stats/prizes | 奖品中奖统计（支持 activity_id 筛选） |
