package migrations

// SQL
const (
	sqlCreateTableKafkaOffsets string = `
-- Таблица для ручного (внешнего) отслеживания смещений потребителей Kafka
CREATE TABLE IF NOT EXISTS kafka_offsets (
    id varchar(50) NOT NULL,
    -- Идентификатор логической группы воркеров
    consumer_group VARCHAR(255) NOT NULL,
    
    -- Топик брокера, из которого вычитываются сообщения
    topic          VARCHAR(255) NOT NULL,
    
    -- Номер конкретной партиции (Partition ID) внутри топика
    partition_id   INT          NOT NULL,
    
    -- Номер последнего УСПЕШНО обработанного и зафиксированного смещения (Offset)
    current_offset BIGINT       NOT NULL,
    
    -- Временная метка последнего обновления записи для аудита и дебага
    updated_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    -- ---------------------------------------------------------------------------
    constraint kafka_offsets_pk primary key (id),
    constraint kafka_offsets_uk unique (consumer_group, topic)
)
`
	sqlDropTableKafkaOffsets string = `
DROP TABLE IF EXISTS kafka_offsets
`
	sqlCreateIndexKafkaOffsets string = `
-- Индекс для мгновенной выборки всех оффсетов конкретного топика при старте или ребалансировке пода.
-- Оптимизирует скорость инициализации метода SetOffset на больших объемах партиций.
CREATE INDEX IF NOT EXISTS idx_kafka_offsets_lookup ON kafka_offsets (consumer_group, topic)
`
	sqlDropIndexKafkaOffsets string = `
drop index idx_kafka_offsets_lookup
`
)
