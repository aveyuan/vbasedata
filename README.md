# vbasedata 基础组件库
封装了一些基础组件，如id生成器、数据库、kafka等
## 组件
- id生成器
- 数据库
- kafka
- 邮箱

```env
# vbasedata 集成测试环境变量
# 测试会通过 TestMain 自动加载本文件（见 env_loader_test.go），直接 go test ./... 即可。
# 真实环境变量优先于本文件；未设置的变量回退到测试代码里的默认值，
# 默认凭据连不上时对应测试会自动 Skip。

# ---------- MySQL（TestNewGorm_MySQL_AutoCreateDatabase）----------
VB_TEST_MYSQL_ADDR=127.0.0.1:3306
VB_TEST_MYSQL_USER=root
VB_TEST_MYSQL_PASS=123456

# ---------- PostgreSQL（TestNewGorm_PG_AutoCreateDatabase）----------
VB_TEST_PG_ADDR=127.0.0.1:5432
VB_TEST_PG_USER=postgres
VB_TEST_PG_PASS=123456
VB_TEST_PG_SSLMODE=disable
# 可选，留空则用系统时区
VB_TEST_PG_TIMEZONE=

# ---------- SMTP（TestEmail_SendMsg）----------
VB_TEST_SMTP_HOST=
VB_TEST_SMTP_PORT=
VB_TEST_SMTP_FROM=
VB_TEST_SMTP_TO=
VB_TEST_SMTP_USER=
VB_TEST_SMTP_PASS=
# 置为 1 或 true 启用隐式 SSL/TLS（如端口 465）
VB_TEST_SMTP_TLS=1

# ---------- Redis（TestNewRedis_LocalEnv）----------
VB_TEST_REDIS_ADDR=127.0.0.1:6379
VB_TEST_REDIS_AUTH=
VB_TEST_REDIS_DB=0

```