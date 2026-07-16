# vbasedata 基础组件库

`vbasedata` 是面向 Go 服务的基础设施组件封装，包含数据库、Redis、验证码、邮件、任务池和雪花 ID 生成器。

## 组件

- `NewGorm`：MySQL、PostgreSQL、SQLite 的 GORM 客户端；连接成功后返回关闭函数。数据库类型仅支持 `mysql`、`pg`、`postgres`、`sqlite`（或空值）。
- `NewRedis`：Redis 单机、集群和哨兵客户端；连接成功后返回关闭函数。
- `NewCaptcha`：数学验证码。`Generate` 只返回 ID 与图片数据；答案保存在 `Store` 中，使用 `Verify` 校验并消费。
- `Cache` / `NewCache`：验证码、登录验证码与登录锁定使用的短期缓存接口；传入 Redis 客户端时使用 Redis，否则回退到带过期时间的 `NewLruCache`。`RedisCache` 与 `LruCache` 都支持原子消费读取和校验。
- `NewPond`：任务池封装。
- `NewEmail`：SMTP 邮件发送。
- `NewIdgenerator`：雪花 ID 生成器。

连接型客户端必须关闭：

```go
db, closeDB, err := vbasedata.NewGorm(config, logger)
if err != nil {
	return err
}
defer closeDB()
_ = db
```

## 测试

默认测试只运行本地单元测试，不会读取 `.env` 或连接外部服务：

```bash
go test ./...
go test -race ./...
go vet ./...
```

MySQL、PostgreSQL、Redis 和 SMTP 测试是显式启用的集成测试。仅在隔离环境中设置凭据，并执行：

```bash
VB_RUN_INTEGRATION=1 go test ./...
```

集成测试启用后才会读取仓库根目录的本地 `.env`；该文件已被 Git 忽略，不能提交真实凭据。