package db_job

import (
	"os"

	"github.com/rs/zerolog/log"
)

func (j *DbJobHandler) RestoreDb(job *DbJob) error {
	if _, err := os.Stat(job.Data.BackupPath); err != nil {
		log.Warn().Err(err).
			Str("BackupPath", job.Data.BackupPath).
			Msg("Failed to restore database")
		return err
	}

	// TODO

	return nil
}
