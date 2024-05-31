package postgres

import (
	"context"
	"time"

	"github.com/EgorTarasov/streaming/orchestrator/pkg/db"
)

type TaskRepo struct {
	db *db.Database
}

func New(database *db.Database) *TaskRepo {
	return &TaskRepo{
		db: database,
	}
}

func (tr *TaskRepo) Create(ctx context.Context, title, rtspUrl string) (int64, error) {
	query := `insert into tasks(title, src, status, type) values($1, $2, 'PENDING'::task_status, 'STREAM'::video_type) returning id;`
	var id int64
	if err := tr.db.Get(ctx, &id, query, title, rtspUrl); err != nil {
		return 0, err
	}
	return id, nil
}

func (tr *TaskRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `update tasks set status = $1::task_status where id = $2;`
	if _, err := tr.db.Exec(ctx, query, status, id); err != nil {
		return err
	}
	return nil
}

func (tr *TaskRepo) UpdateSplitFrames(ctx context.Context, id int64, status string, frames int64) error {
	query := `update tasks set status = $1::task_status, split_frames = $2 where id = $3;`
	if _, err := tr.db.Exec(ctx, query, status, frames, id); err != nil {
		return err
	}
	return nil
}

func (tr *TaskRepo) UpdatePredictedFrames(ctx context.Context, id int64, frames int64) error {
	query := `update tasks set predicted_frames = $1 where id = $2;`
	if _, err := tr.db.Exec(ctx, query, frames, id); err != nil {
		return err
	}
	return nil
}

// TODO: move to models
type TaskStatusDao struct {
	VideoId         int64     `db:"id"`
	SplitFrames     int64     `db:"split_frames"`
	ProcessedFrames int64     `db:"predicted_frames"`
	Status          string    `db:"status"`
	CreatedAt       time.Time `db:"created_at"`
}

func (tr *TaskRepo) GetTaskStatus(ctx context.Context, id int64) (TaskStatusDao, error) {
	query := `select id, split_frames, predicted_frames, status, created_at from tasks where id = $1;`
	var task TaskStatusDao
	if err := tr.db.Get(ctx, &task, query, id); err != nil {
		return task, err
	}
	return task, nil
}

func (tr *TaskRepo) SaveResultVideo(ctx context.Context, videoId int64, s3Url string) (int64, error) {
	var newId int64
	query := `insert into result_video(task_id, video_url) values ($1, $2) returning id;`
	if err := tr.db.Get(ctx, &newId, query, videoId, s3Url); err != nil {
		return newId, err
	}
	return newId, nil
}

func (tr *TaskRepo) GetResultVideo(ctx context.Context, videoId int64) (string, error) {
	var s3Url string
	query := `select video_url from result_video where task_id = $1 order by created_at desc limit 1;`
	if err := tr.db.Get(ctx, &s3Url, query, videoId); err != nil {
		return s3Url, err
	}
	return s3Url, nil
}
