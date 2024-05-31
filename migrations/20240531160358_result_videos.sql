-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
create table if not exists result_video(
    id bigserial primary key,
    task_id bigint references tasks(id),
    video_url text not null,
    created_at timestamp not null default now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
drop table result_frames;
-- +goose StatementEnd
