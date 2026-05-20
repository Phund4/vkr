package services

// Лимиты, разделители ключей и таймауты ingest-слоя.
const (
	maxIngestBodyBytes    = 8 << 20
	maxLabelLen           = 128
	publishRequestTimeout = 8 // секунд — публикация в Kafka

	// congestionPairKeySep разделитель segment и camera в карте интервалов congestion.
	congestionPairKeySep = "\x00"

	// rawMLJSONEmpty значение RawML при отсутствии тела ml.
	rawMLJSONEmpty = "{}"
	// incidentLabelCrash подстрока класса инцидента для правила алерта.
	incidentLabelCrash = "crash"

	// ingestMetricStageReadBody стадия ошибки: чтение тела.
	ingestMetricStageReadBody = "read_body"
	// ingestMetricStageJSONDecode стадия: разбор JSON.
	ingestMetricStageJSONDecode = "json_decode"
	// ingestMetricStageValidate стадия: валидация полей.
	ingestMetricStageValidate = "validate"
	// ingestMetricStageJSONEncode стадия: сериализация persist JSON.
	ingestMetricStageJSONEncode = "json_encode"

	// kafkaPublishErrorStageWrite стадия ошибки записи в Kafka.
	kafkaPublishErrorStageWrite = "write"
)
