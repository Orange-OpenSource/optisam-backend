-- name: PublishToDLQ :exec
INSERT INTO dead_letter_queue(
    topic,
    no_of_retries,
    message
) VALUES(
    $1,
    $2,
    $3
);
