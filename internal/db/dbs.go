package db

import (
	"context"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/jackc/pgx/v5"
)

func ListDatabases(q *Queries) ([]*proto.Database, error) {
	dbs, err := q.ListDbs(context.Background())
	if err != nil {
		if err == pgx.ErrNoRows {
			return []*proto.Database{}, nil
		}
		return nil, err
	}

	databases := make([]*proto.Database, len(dbs))
	for i, db := range dbs {
		databases[i] = ConvertDbToDatabase(&db)
	}
	return databases, nil
}
