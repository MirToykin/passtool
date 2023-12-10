package cmd

import (
	"fmt"
	"github.com/MirToykin/passtool/internal/config"
	"github.com/MirToykin/passtool/internal/crypto"
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	passGenerator "github.com/sethvargo/go-password/password"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// requestUniqueLoginForService request login from user. If login already exists for the given service - retries.
func requestUniqueLoginForService(
	account *models.Account,
	service models.Service,
	printer Printer,
	db *gorm.DB,
) (string, error) {
	for {
		login := cli.GetUserInput("Enter login: ", printer)

		var count int64
		err := account.FindByLoginAndServiceID(db, login, service.ID).Count(&count).Error
		if err != nil {
			return "", fmt.Errorf("faild to request unique login: %w", err)
		}

		if count > 0 {
			log.Printf(
				"Account with login %q at %q already exists, to update it use %q command. Use another login.",
				login,
				service.Name,
				updateCmd.Use,
			)
		} else {
			return login, nil
		}
	}
}

// getSecret returns secret given by user (handle possible errors)
func getSecret(secretName string, confirm bool, printer Printer) string {
	postfix := ""
	if confirm {
		postfix = " again"
	}

	secret, err := cli.GetSensitiveUserInput(fmt.Sprintf("Enter %s%s: ", secretName, postfix), printer)
	checkSimpleErrorWithDetails(err, fmt.Sprintf("unable to get %s", secretName), printer)
	return secret
}

// getSecretWithConfirmation handles getting pass phrase with confirmation
func getSecretWithConfirmation(secretName string, retryMsg string, printer Printer) string {
	for {
		pass1 := getSecret(secretName, false, printer)
		pass2 := getSecret(secretName, true, printer)

		if pass1 != pass2 {
			fmt.Println(retryMsg)
		} else {
			return pass1
		}
	}
}

// checkSimpleErrorWithDetails handles general error, and prints message with error details to the console
func checkSimpleErrorWithDetails(err error, msg string, p Printer) {
	if err != nil {
		p.ErrorWithExit("%s: %v", msg, err)
	}
}

// checkSimpleError handles general error and prints message to the console.
func checkSimpleError(err error, msg string, p Printer) {
	if err != nil {
		p.ErrorWithExit(msg)
	}
}

type GenSettings interface {
	GetLength() int
	GetNumDigits() int
	GetNumSymbols() int
	GetNoUpper() bool
	GetAllowRepeat() bool
}

// encryptPassword sets encrypted password and salt for given Password instance
func encryptPassword(
	password *models.Password,
	userPassword, secret string,
	keyLen int,
	genSettings GenSettings,
) error {
	salt, err := passGenerator.Generate(
		genSettings.GetLength(),
		genSettings.GetNumDigits(),
		genSettings.GetNumSymbols(),
		genSettings.GetNoUpper(),
		genSettings.GetAllowRepeat())

	if err != nil {
		return fmt.Errorf("unable to get salt: %w", err)
	}

	key := crypto.DeriveKey(secret, salt, keyLen)

	encryptedPassword, err := crypto.Encrypt(key, userPassword)
	if err != nil {
		return fmt.Errorf("unable to encrypt the password: %w", err)
	}

	password.Encrypted = encryptedPassword
	password.Salt = salt
	return nil
}

// PrintServiceRequirements prints the information for service to be able to work
func PrintServiceRequirements(cfg *config.Config, printer Printer) {
	fmt.Println()
	printer.Infoln("For the app to work you need to add the following environment variables:")
	for _, ev := range cfg.GetRequiredEnvVars() {
		fmt.Println(fmt.Sprintf("  %q - %s", ev.Name, ev.Description))
	}

	fmt.Println()

	printer.Infoln("You might also want to add the following optional environment variables:")
	for _, ev := range cfg.GetOptionalEnvVars() {
		fmt.Println(fmt.Sprintf("  %q - %s", ev.Name, ev.Description))
	}
	fmt.Println()
}

// copyFile copies file from source to destination
func copyFile(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("unable to open source file: %w", err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("unable to create destination file: %w", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("unable to copy file: %w", err)
	}

	return nil
}

// checkIfBackupNeeded returns true if backup needed otherwise false
func checkIfBackupNeeded(passwordID, backupIndex uint) bool {
	return passwordID%backupIndex == 0
}

// createBackup creates storage backup
func createBackup(
	wg *sync.WaitGroup,
	cfg *config.Config,
	errChan chan<- error,
	printer Printer) {
	defer wg.Done()
	printer.Simpleln("Creating backup...")

	err := copyFile(cfg.StoragePath, cfg.GetBackupFilePath())
	if err != nil {
		errChan <- fmt.Errorf("unable to create backup: %w", err)
		return
	}

	errChan <- nil
}

// clearUnnecessaryBackups clears outdated backups
func clearUnnecessaryBackups(
	wg *sync.WaitGroup,
	errChan chan<- error,
	conf *config.Config,
	printer Printer,
) {
	defer wg.Done()
	errTemplate := "unable to clear backup: %w"
	files, err := filepath.Glob(conf.GetBackupFilePathMask())
	if err != nil {
		errChan <- fmt.Errorf(errTemplate, err)
		return
	}

	if len(files) <= int(conf.BackupCountToStore) {
		errChan <- nil
		return
	}

	printer.Simpleln("clearing unnecessary backups...")

	type fileInfo struct {
		path    string
		created int64
	}
	var fileInfoData []fileInfo

	for _, file := range files {
		fInfo, err := os.Stat(file)
		if err != nil {
			errChan <- fmt.Errorf(errTemplate, err)
			return
		}

		fileInfoData = append(fileInfoData, fileInfo{
			path:    file,
			created: fInfo.ModTime().Unix(),
		})
	}

	sort.Slice(fileInfoData, func(i, j int) bool {
		return fileInfoData[i].created > fileInfoData[j].created
	})

	for _, file := range fileInfoData[conf.BackupCountToStore:] {
		err = os.Remove(file.path)
		if err != nil {
			printer.Warning("unable to remove unnecessary backup file: %s", file.path)
		}
	}

	errChan <- nil
}

func genericAdd(
	operation string,
	db *gorm.DB,
	printer Printer,
	conf *config.Config,
	getPassword func() string,
) {
	var service models.Service
	serviceName := cli.GetUserInput("Enter service name: ", printer)
	err := service.FetchOrCreate(db, serviceName)
	checkSimpleErrorWithDetails(err, operation, printer)

	var account models.Account
	login, err := requestUniqueLoginForService(&account, service, printer, db)
	checkSimpleErrorWithDetails(err, operation, printer)

	account.Service = service
	account.Login = login

	var password models.Password
	userPassword := getPassword()
	secretKey := getSecretWithConfirmation("secret key", "Secret keys are not equal", printer)

	err = encryptPassword(&password, userPassword, secretKey, conf.SecretKeyLength, conf.PasswordSettings)
	checkSimpleErrorWithDetails(err, operation, printer)

	err = account.SaveWithPassword(db, &password)
	checkSimpleErrorWithDetails(err, operation, printer)

	wg := sync.WaitGroup{}
	errChan := make(chan error, 2)

	if checkIfBackupNeeded(password.ID, conf.BackupIndex) {
		wg.Add(2)
		go createBackup(&wg, conf, errChan, printer)
		go clearUnnecessaryBackups(&wg, errChan, conf, printer)
	}

	printer.Success("Successfully added password for account with login %q at %q", login, serviceName)

	wg.Wait()
	close(errChan)

	for err = range errChan {
		if err != nil {
			printer.Warning("failed to handle backup: %v", err)
		}
	}
}
