package config

import (
	"os"
)

type Config struct {
	StoragePath      string
	BackupIndex      uint64
	SecretKeyLength  int
	PasswordSettings GeneratorSettings
	SaltSettings     GeneratorSettings
	EnvVariables     []EnvVar
}

func (c Config) IsValid() bool {
	for _, ev := range c.EnvVariables {
		if ev.Required && ev.Value == "" {
			return false
		}
	}

	return true
}

func (c Config) filterVars(required bool) (requiredVars []EnvVar) {
	for _, ev := range c.EnvVariables {
		if ev.Required == required {
			requiredVars = append(requiredVars, ev)
		}
	}

	return
}

func (c Config) GetRequiredEnvVars() []EnvVar {
	return c.filterVars(true)
}

func (c Config) GetOptionalEnvVars() []EnvVar {
	return c.filterVars(false)
}

type GeneratorSettings struct {
	Length      int
	NumDigits   int
	NumSymbols  int
	NoUpper     bool
	AllowRepeat bool
}

func Load() *Config {
	storageVar.Value = ensureTrailingSlash(os.Getenv(storageEnv)) + storageFileName
	backupIndexVar.Value = os.Getenv(backupIndexEnv)
	backupIndex := backupIndexVar.IntVal()

	if backupIndex == 0 {
		backupIndex = defaultBackupIndex
	}

	var cfg = Config{
		StoragePath:     storageVar.Value,
		BackupIndex:     backupIndex,
		SecretKeyLength: 32,
		PasswordSettings: GeneratorSettings{
			Length:      12,
			NumDigits:   3,
			NumSymbols:  3,
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
		EnvVariables: []EnvVar{storageVar, backupIndexVar},
	}

	return &cfg
}
