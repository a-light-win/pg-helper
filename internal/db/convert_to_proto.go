package db

import (
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (db *Db) ToProto() *proto.Database {
	if db == nil {
		return nil
	}
	return &proto.Database{
		Name:        db.Name,
		Owner:       db.Owner,
		CreatedAt:   pgTimestampToProto(db.CreatedAt),
		UpdatedAt:   pgTimestampToProto(db.UpdatedAt),
		ExpiredAt:   pgTimestampToProto(db.ExpiredAt),
		MigrateFrom: db.MigrateFrom,
		MigrateTo:   db.MigrateTo,
		Status:      db.Status,
		Stage:       db.Stage,
		ErrorMsg:    db.ErrorMsg,
	}
}

func pgTimestampToProto(ts pgtype.Timestamp) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}
	return timestamppb.New(ts.Time)
}
