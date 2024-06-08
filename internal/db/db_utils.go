package db

import "github.com/a-light-win/pg-helper/pkg/proto"

func (d *Db) IsAlreadyIdle() bool {
	return (d.Stage == proto.DbStage_Idle && d.Status == proto.DbStatus_Done) ||
		d.Stage == proto.DbStage_DropDatabase
}

func (d *Db) IsNotExist() bool {
	return d.Stage == proto.DbStage_None
}

func (d *Db) IsReadyToUse() bool {
	return d.Stage == proto.DbStage_ReadyToUse && d.Status == proto.DbStatus_Done
}

func (d *Db) CanRestore() bool {
	return d.Status == proto.DbStatus_Done &&
		(d.Stage == proto.DbStage_CreateDatabase ||
			d.Stage == proto.DbStage_RestoreDatabase)
}

func (d *Db) ShouldBackup() bool {
	return (d.Stage == proto.DbStage_CreateDatabase && d.Status == proto.DbStatus_Done) ||
		(d.Stage == proto.DbStage_BackupDatabase && d.Status == proto.DbStatus_Failed)
}
