package vbasedata

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestMain 只在显式启用集成测试时读取本地 .env，避免本机凭据和外部服务
// 影响默认的单元测试。
func TestMain(m *testing.M) {
	if integrationEnabled() {
		if err := loadDotEnv(".env"); err != nil {
			fmt.Fprintf(os.Stderr, "warning: load .env: %v\n", err)
		}
	}
	os.Exit(m.Run())
}

func integrationEnabled() bool {
	return os.Getenv("VB_RUN_INTEGRATION") == "1"
}

func requireIntegration(t *testing.T) {
	t.Helper()
	if !integrationEnabled() {
		t.Skip("集成测试未启用；设置 VB_RUN_INTEGRATION=1 后运行")
	}
}

// loadDotEnv 解析简单的 KEY=VALUE 形式的 .env 文件。
// 规则：忽略空行与 # 注释；可选地去除值两侧的引号；
// 已存在的环境变量优先，不会被 .env 覆盖（真实环境 > .env）。
// 文件不存在视为正常（返回 nil）。
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return scanner.Err()
}
