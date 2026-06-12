package vbasedata

import (
	"github.com/yitter/idgenerator-go/idgen"
)

type Idgenerator struct {
	gen *idgen.DefaultIdGenerator
}

// NewIdgenerator 创建一个独立的雪花ID生成器实例。
// 使用 NewDefaultIdGenerator 而非全局的 idgen.SetIdGenerator，
// 这样多个实例（不同 workId）之间互不影响。
func NewIdgenerator(workId uint16) *Idgenerator {
	options := idgen.NewIdGeneratorOptions(workId)
	options.WorkerIdBitLength = 4 // 默认值6，限定 WorkerId 最大值为2^4-1，即最多支持16个节点。
	options.SeqBitLength = 6      // 默认值6，限制每毫秒生成的ID个数。若生成速度超过5万个/秒，建议加大 SeqBitLength 到 10。
	return &Idgenerator{
		gen: idgen.NewDefaultIdGenerator(options),
	}
}

func (t *Idgenerator) NextId() int64 {
	return t.gen.NewLong()
}
