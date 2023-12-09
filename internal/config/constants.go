package config

const (
	// required env vars
	storageEnv = "PASSTOOL_STORAGE_PATH"

	// optional env vars
	backupIndexEnv = "PASSTOOL_BACKUP_INDEX"
	backupCountEnv = "PASSTOOL_BACKUP_COUNT"

	// Defaults
	defaultBackupIndex = 5
	defaultBackupCount = 5

	//Other
	storageFileName               = "passtool_storage.db"
	storageBackupFileNameTemplate = "%v.passtool_backup.db"
)
