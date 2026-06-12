package vbasedata

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

// TestMain 在运行测试前自动加载包目录下的 .env 文件，
// 这样无需手动 source 即可让集成测试读取到环境变量。
// 注意：整个测试包只能有一个 TestMain。
func TestMain(m *testing.M) {
	loadDotEnv(".env")
	os.Exit(m.Run())
}

// loadDotEnv 解析简单的 KEY=VALUE 形式的 .env 文件。
// 规则：忽略空行与 # 注释；可选地去除值两侧的引号；
// 已存在的环境变量优先，不会被 .env 覆盖（真实环境 > .env）。
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // 没有 .env 就跳过
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}
		// 去除值两侧可选的引号
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') ||
				(val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if _, exists := os.LookupEnv(key); exists {
			continue // 真实环境变量优先
		}
		_ = os.Setenv(key, val)
	}
}
