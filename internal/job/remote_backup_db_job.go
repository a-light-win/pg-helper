package job

func (j *DbJob) execRemoteBackup() {
	remoteUrl := j.DbConfig.RemoteHelperUrl(j.DbTask.Data.BackupVersion)
	remoteUrl += "/api/v1/db/backup"

	request := dto.BackupRequest{}
}
