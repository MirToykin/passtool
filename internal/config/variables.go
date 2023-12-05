package config

import (
	"fmt"
	"log"
	"strconv"
)

type VarType int

const (
	EnvStr VarType = iota
	EnvInt
)

type EnvVar struct {
	Name        string
	Description string
	Value       string
	Type        VarType
	Required    bool
}

func (ev EnvVar) IntVal() uint64 {
	if ev.Type != EnvInt || ev.Value == "" {
		return 0
	}

	intVal, err := strconv.ParseUint(ev.Value, 10, 64)
	if err != nil {
		log.Fatalf("can't convert %q environment variable to int", ev.Name)
	}

	return intVal
}

var storageVar = EnvVar{
	Name:        storageEnv,
	Description: "Path to a directory where your encrypted data will be stored, e.g. /Users/me/passtool_storage.db",
	Type:        EnvStr,
	Required:    true,
}

var backupIndexVar = EnvVar{
	Name:        backupIndexEnv,
	Description: fmt.Sprintf("Do DB backup per each N passwords, by default its value is %d", defaultBackupIndex),
	Type:        EnvInt,
	Required:    false,
}
