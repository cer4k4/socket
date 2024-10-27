CREATE TABLE IF NOT EXISTS messages (
                                        chat_id bigint PRIMARY KEY,
                                        profit int,
                                        failed_update_attempts int
);
