package db

import (
	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertDbToDatabase(db *Db) *proto.Database {
	return &proto.Database{
		Name:        db.Name,
		Owner:       db.Owner,
		CreatedAt:   convertPgTimestampToProtoTimestamp(db.CreatedAt),
		UpdatedAt:   convertPgTimestampToProtoTimestamp(db.UpdatedAt),
		ExpiredAt:   convertPgTimestampToProtoTimestamp(db.ExpiredAt),
		MigrateFrom: db.MigrateFrom,
		MigrateTo:   db.MigrateTo,
		Status:      db.Status,
	}
}

func convertPgTimestampToProtoTimestamp(ts pgtype.Timestamp) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}
	return timestamppb.New(ts.Time)
}
