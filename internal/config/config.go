package config

import (
	"fmt"
	"path/filepath"
	"time"
)

type Config struct {
	BasePath               string
	StoragePath            string
	BackupFilenameTemplate string
	BackupIndex            uint
	BackupCountToStore     uint
	SecretKeyLength        int
	MinPasswordLength      int
	MaxPasswordLength      int
	PasswordSettings       GeneratorSettings
	SaltSettings           GeneratorSettings
	EnvVariables           []EnvVar
}

// IsValid checks if config is valid
func (c Config) IsValid() bool {
	for _, ev := range c.EnvVariables {
		if ev.Required && ev.Value == "" {
			return false
		}
	}

	return true
}

// filterVars returns EnvVar objects filtered by Required key
func (c Config) filterVars(required bool) (requiredVars []EnvVar) {
	for _, ev := range c.EnvVariables {
		if ev.Required == required {
			requiredVars = append(requiredVars, ev)
		}
	}

	return
}

// GetRequiredEnvVars return EnvVar objects for required ENV variables
func (c Config) GetRequiredEnvVars() []EnvVar {
	return c.filterVars(true)
}

// GetOptionalEnvVars return EnvVar objects for optional ENV variables
func (c Config) GetOptionalEnvVars() []EnvVar {
	return c.filterVars(false)
}

// GetBackupFilePath returns path to directory where backups are stored
func (c Config) GetBackupFilePath() string {
	fileName := fmt.Sprintf(c.BackupFilenameTemplate, time.Now().Unix())
	return filepath.Join(c.BasePath, fileName)
}

// GetBackupFilePathMask returns generic mask for backup file name
func (c Config) GetBackupFilePathMask() string {
	return filepath.Join(c.BasePath, fmt.Sprintf(c.BackupFilenameTemplate, "*"))
}

type GeneratorSettings struct {
	Length      int
	NumDigits   int
	NumSymbols  int
	NoUpper     bool
	AllowRepeat bool
}

// GetLength returns length of generated password
func (gs GeneratorSettings) GetLength() int {
	return gs.Length
}

// GetNumDigits returns the number of digits to use in generated password
func (gs GeneratorSettings) GetNumDigits() int {
	return gs.NumDigits
}

// GetNumSymbols returns the number of spec symbols to use in generated password
func (gs GeneratorSettings) GetNumSymbols() int {
	return gs.NumSymbols
}

// GetNoUpper whether to allow lowercase characters in generated password
func (gs GeneratorSettings) GetNoUpper() bool {
	return gs.NoUpper
}

// GetAllowRepeat whether to allow repeated characters in generated password
func (gs GeneratorSettings) GetAllowRepeat() bool {
	return gs.AllowRepeat
}

// Load creates and returns pointer to Config
func Load() *Config {
	environment.loadVars()
	storageDir := environment.getStorage()
	return &Config{
		BasePath:               storageDir,
		StoragePath:            filepath.Join(storageDir, storageFileName),
		BackupFilenameTemplate: storageBackupFileNameTemplate,
		BackupIndex:            environment.getBackupIndex(),
		BackupCountToStore:     environment.getBackupCount(),
		SecretKeyLength:        32,
		MinPasswordLength:      6,
		MaxPasswordLength:      100,
		PasswordSettings: GeneratorSettings{
			Length:      int(environment.getDefaultPasswordLength()),
			NumDigits:   4,
			NumSymbols:  4,
			NoUpper:     false,
			AllowRepeat: false,
		},
		SaltSettings: GeneratorSettings{
			Length:      32,
			NumDigits:   7,
			NumSymbols:  8,
			NoUpper:     false,
			AllowRepeat: false,
		},
		EnvVariables: environment.getVars(),
	}
}
