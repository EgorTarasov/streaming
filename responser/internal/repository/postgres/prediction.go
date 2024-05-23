package postgres

import (
	"context"

	"github.com/EgorTarasov/streaming/responser/pkg/infrastructure/postgres"
)

type PredictionRepository struct {
	db *postgres.Database
}

func NewPredictionRepository(db *postgres.Database) *PredictionRepository {
	return &PredictionRepository{db: db}
}

func (pr *PredictionRepository) SavePredictionResult(ctx context.Context, videoId int64, frameId int64, result interface{}, s3Key string) (int64, error) {
	query := `insert into result_frames (fk_task, frame, metadata, s3_object_key) values ($1, $2, $3, $4) returning id;`
	var id int64
	if err := pr.db.Get(ctx, &id, query, videoId, frameId, result, s3Key); err != nil {
		return 0, err
	}
	return id, nil
}
