CREATE TABLE IF NOT EXISTS messages (
    id UUID,
    entity_type INT,
    operation INT,
    data VARCHAR,
    chat_id BIGINT,
    profit INT,
    failed_update_attempts INT,
    PRIMARY KEY ((chat_id), id)
)
WITH CLUSTERING ORDER BY (id DESC);