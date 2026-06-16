package kafkapub

import "time"

const (
	// writerBatchTimeout задержка батчинга записи Kafka writer.
	writerBatchTimeout = 10 * time.Millisecond
)
