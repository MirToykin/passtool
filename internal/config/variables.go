package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type VarType int

const (
	EnvStr VarType = iota
	EnvInt
)

type EnvVar struct {
	Name            string
	Description     string
	Value           string
	DefaultIntValue uint
	DefaultStrValue string
	Type            VarType
	Required        bool
}

// intVal casts EnvVar.Value to uint64 and returns it. If fails then stops the execution with log.
func (ev EnvVar) intVal() uint {
	if ev.Type != EnvInt {
		return 0
	}

	if ev.Value == "" {
		return ev.DefaultIntValue
	}

	intVal, err := strconv.ParseUint(ev.Value, 10, 64)
	if err != nil {
		log.Fatalf("can't convert %q environment variable to int", ev.Name)
	}

	return uint(intVal)
}

func (ev EnvVar) stringVal() string {
	if ev.Type != EnvStr {
		return ""
	}

	if ev.Value == "" {
		return ev.DefaultStrValue
	}

	return ev.Value
}

var storageVar = EnvVar{
	Name: storageEnv,
	Description: `Path to a directory where your data will be stored, e.g. /Users/me/passtool.
			    Passwords data keeps encrypted.`,
	Type:     EnvStr,
	Required: true,
}

var backupIndexVar = EnvVar{
	Name:            backupIndexEnv,
	Description:     fmt.Sprintf("Do DB backup per each N passwords, by default its value is %d", defaultBackupIndex),
	Type:            EnvInt,
	Required:        false,
	DefaultIntValue: defaultBackupIndex,
}

var backupCountVar = EnvVar{
	Name:            backupCountEnv,
	Description:     fmt.Sprintf("Count of backups to store, by default %d", defaultBackupCount),
	Type:            EnvInt,
	Required:        false,
	DefaultIntValue: defaultBackupCount,
}

var defaultPasswordLengthVar = EnvVar{
	Name:            defaultPasswordLengthEnv,
	Description:     fmt.Sprintf("Default generated password length, if not set equals %d", defaultPasswordLength),
	Type:            EnvInt,
	Required:        false,
	DefaultIntValue: defaultPasswordLength,
}

type Environment struct {
	storage               *EnvVar
	backupIndex           *EnvVar
	backupCount           *EnvVar
	defaultPasswordLength *EnvVar
	loaded                bool
	vars                  []*EnvVar
}

func (env *Environment) loadVars() {
	for _, v := range env.vars {
		v.Value = os.Getenv(v.Name)
	}

	env.loaded = true
}

func (env *Environment) getVars() []EnvVar {
	env.mustBeLoaded()

	var loadedVars []EnvVar
	for _, v := range env.vars {
		loadedVars = append(loadedVars, *v)
	}

	return loadedVars
}

func (env *Environment) getStorage() string {
	env.mustBeLoaded()
	return env.storage.stringVal()
}

func (env *Environment) getBackupIndex() uint {
	env.mustBeLoaded()
	return env.backupIndex.intVal()
}

func (env *Environment) getBackupCount() uint {
	env.mustBeLoaded()
	return env.backupCount.intVal()
}

func (env *Environment) getDefaultPasswordLength() uint {
	env.mustBeLoaded()
	return env.defaultPasswordLength.intVal()
}

func (env *Environment) mustBeLoaded() {
	if !env.loaded {
		log.Fatal("Environment is not loaded")
	}
}

var environment = Environment{
	loaded:                false,
	storage:               &storageVar,
	backupIndex:           &backupIndexVar,
	backupCount:           &backupCountVar,
	defaultPasswordLength: &defaultPasswordLengthVar,
	vars:                  []*EnvVar{&storageVar, &backupIndexVar, &backupCountVar, &defaultPasswordLengthVar},
}
