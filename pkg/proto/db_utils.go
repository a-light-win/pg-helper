package proto

func (d *Database) IsAlreadyIdle() bool {
	return d != nil &&
		(d.Stage == DbStage_DropDatabase ||
			(d.Stage == DbStage_Idle && d.Status == DbStatus_Done))
}

func (d *Database) IsNotExist() bool {
	return d == nil || d.Stage == DbStage_None
}

func (d *Database) IsReadyToUse() bool {
	return d != nil &&
		d.Stage == DbStage_ReadyToUse &&
		d.Status == DbStatus_Done
}

func (d *Database) IsFailed() bool {
	return d != nil && d.Status == DbStatus_Failed
}

func (d *Database) IsReadyToMigrate() bool {
	return d.IsNotExist() || d.IsAlreadyIdle()
}

func (d *Database) IsSynced() bool {
	return d != nil &&
		d.Status == DbStatus_Done &&
		(d.Stage == DbStage_ReadyToUse ||
			d.Stage == DbStage_Idle ||
			d.Stage == DbStage_DropDatabase)
}
