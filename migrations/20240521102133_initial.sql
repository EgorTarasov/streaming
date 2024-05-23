-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- PENDING = 0;
-- PROCESSING = 1;
-- DONE = 2;
-- ERROR = 3;
CREATE TYPE task_status AS ENUM ('PENDING', 'PROCESSING', 'DONE', 'ERROR');
CREATE TYPE video_type AS ENUM ('STREAM', 'FILE');
create table if not exists tasks (
    id bigserial primary key,
    title text not null,
    type video_type not null,
    src text not null,
    -- ссылка на поток или файл
    split_frames bigint not null default 0,
    predicted_frames bigint not null default 0,
    status task_status default 'PENDING'::task_status,
    created_at timestamp default now()
);
create table if not exists result_frames(
    id bigserial primary key,
    fk_task bigint references tasks(id) not null,
    frame bigint not null,
    metadata json default '{}'::JSON,
    s3_object_key text not null
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
drop table result_frames;
drop table tasks;
drop type task_status;
drop type video_type;
-- +goose StatementEnd