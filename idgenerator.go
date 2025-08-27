package vbasedata

import (
	"github.com/yitter/idgenerator-go/idgen"
)

type Idgenerator struct {
}

func NewIdgenerator(workId uint16) *Idgenerator {
	var options = idgen.NewIdGeneratorOptions(workId)
	options.WorkerIdBitLength = 4 // 默认值6，限定 WorkerId 最大值为2^6-1，即默认最多支持64个节点。
	options.SeqBitLength = 6      // 默认值6，限制每毫秒生成的ID个数。若生成速度超过5万个/秒，建议加大 SeqBitLength 到 10。
	idgen.SetIdGenerator(options)
	return &Idgenerator{}
}

func (t *Idgenerator) NextId() int64 {
	return idgen.NextId()
}
