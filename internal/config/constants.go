package config

const (
	// required env vars
	storageEnv = "PASSTOOL_STORAGE_PATH"

	// optional env vars
	backupIndexEnv           = "PASSTOOL_BACKUP_INDEX"
	backupCountEnv           = "PASSTOOL_BACKUP_COUNT"
	defaultPasswordLengthEnv = "PASSTOOL_DEFAULT_PASSWORD_LENGTH"

	// Defaults
	defaultBackupIndex    = 5
	defaultBackupCount    = 5
	defaultPasswordLength = 12

	//Other
	storageFileName               = "passtool_storage.db"
	storageBackupFileNameTemplate = "%v.passtool_backup.db"
)
