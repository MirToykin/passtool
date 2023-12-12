package cmd

import (
	"errors"
	"fmt"
	"github.com/MirToykin/passtool/internal/config"
	"github.com/MirToykin/passtool/internal/crypto"
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	passGenerator "github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
				"Account with login %q at %q already exists, to update it use the %q command. Use another login.",
				login,
				service.Name,
				"set",
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
	deps AppDependencies,
	getPassword func() string,
) {
	var service models.Service
	serviceName := cli.GetUserInput("Enter service name: ", deps.printer)
	err := service.FetchOrCreate(deps.db, serviceName)
	checkSimpleErrorWithDetails(err, operation, deps.printer)

	var account models.Account
	login, err := requestUniqueLoginForService(&account, service, deps.printer, deps.db)
	checkSimpleErrorWithDetails(err, operation, deps.printer)

	account.Service = service
	account.Login = login

	var password models.Password
	userPassword := getPassword()
	secretKey := getSecretWithConfirmation("secret key", "Secret keys are not equal", deps.printer)

	err = encryptPassword(&password, userPassword, secretKey, deps.config.SecretKeyLength, deps.config.PasswordSettings)
	checkSimpleErrorWithDetails(err, operation, deps.printer)

	err = account.SaveWithPassword(deps.db, &password)
	checkSimpleErrorWithDetails(err, operation, deps.printer)

	wg := sync.WaitGroup{}
	errChan := make(chan error, 2)

	if checkIfBackupNeeded(password.ID, deps.config.BackupIndex) {
		wg.Add(2)
		go createBackup(&wg, deps.config, errChan, deps.printer)
		go clearUnnecessaryBackups(&wg, errChan, deps.config, deps.printer)
	}

	deps.printer.Success("Successfully added password for account with login %q at %q", login, serviceName)

	wg.Wait()
	close(errChan)

	for err = range errChan {
		if err != nil {
			deps.printer.Warning("failed to handle backup: %v", err)
		}
	}
}

// requestExistingService queries for an existing service name until it gets one.
// If succeeds - loads service with its accounts
func requestExistingService(db *gorm.DB, service *models.Service, p Printer) error {
	for {
		serviceName := cli.GetUserInput("Enter service name: ", p)

		ok, err := fetchServiceWithAccounts(db, service, serviceName)
		if err != nil {
			return fmt.Errorf("failed to request existing service: %w", err)
		}

		if !ok {
			p.Warning("Service with name %q not found, try again", serviceName)
		} else {
			return nil
		}
	}
}

// fetchServiceWithAccounts tries to fetch service
func fetchServiceWithAccounts(db *gorm.DB, service *models.Service, serviceName string) (bool, error) {
	err := service.FetchByName(db, serviceName, true)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("unable to fetch service: %w", err)
	}

	return true, nil
}

// requestExistingModel requests name or serial number from user. If it doesn't exist for the given map or slice - retries.
// If succeeds - returns pointer to a given object
func requestExistingModel[M any](
	mMap map[int]M,
	mSlice []M,
	getModelValue func(m M) string,
	strIdentifier string,
	printer Printer,
) *M {
	for {
		identifier := cli.GetUserInput(fmt.Sprintf("Enter %s or serial number: ", strIdentifier), printer)
		var name string

		num, err := strconv.Atoi(identifier)
		if err != nil {
			name = identifier
		} else {
			model, found := mMap[num]
			if found {
				return &model
			} else {
				name = identifier
			}
		}

		for _, model := range mSlice {
			if getModelValue(model) == name {
				return &model
			}
		}

		printer.Warning(
			"Incorrect %s %q, use another one or the correct serial number.",
			strIdentifier,
			name,
		)
	}
}

// printSortedMap prints map key and corresponding string value retrieved from map by key
func printSortedMap[M any](target map[int]M, getStrVal func(target map[int]M, key int) string) {
	keys := make([]int, 0, len(target))

	for key := range target {
		keys = append(keys, key)
	}

	sort.Ints(keys)
	for _, key := range keys {
		fmt.Println(key, getStrVal(target, key))
	}

	fmt.Println()
}

func genericGet(
	operation string,
	db *gorm.DB,
	printer Printer,
	handler func(p models.Password),
) {
	var service *models.Service
	var count int64

	err := service.List(db).Count(&count).Error
	checkSimpleErrorWithDetails(err, "Unable to check service existence", printer)
	if count == 0 {
		printer.Infoln("There are no added services yet")
		os.Exit(0)
	}

	servicesMap, err := service.GetMap(db)
	checkSimpleErrorWithDetails(err, operation, printer)
	var servicesSlice []models.Service
	for _, s := range servicesMap {
		servicesSlice = append(servicesSlice, s)
	}
	printer.Header("The following services were created:")
	printSortedMap(servicesMap, func(sMap map[int]models.Service, key int) string {
		return sMap[key].Name
	})

	service = requestExistingModel(
		servicesMap,
		servicesSlice,
		func(s models.Service) string {
			return s.Name
		},
		"service name",
		printer,
	)

	err = service.LoadAccounts(db)
	checkSimpleErrorWithDetails(err, operation, printer)

	if len(service.Accounts) == 0 {
		printer.Simpleln("Accounts for service %q not found", service.Name)
		os.Exit(0)
	}

	printer.Header("Service %q has accounts with the following logins:", service.Name)
	accountsMap := service.GetAccountsMap()
	printSortedMap(accountsMap, func(aMap map[int]models.Account, key int) string {
		return aMap[key].Login
	})

	account := requestExistingModel(
		accountsMap,
		service.Accounts,
		func(acc models.Account) string {
			return acc.Login
		},
		"login",
		printer,
	)
	err = account.LoadPassword(db)
	checkSimpleErrorWithDetails(err, operation, printer)

	handler(account.Password)
}

func getDecryptedPasswordWithRetry(
	password models.Password,
	keyLen int,
	maxRetries int,
	printer Printer,
) (string, error) {
	tryCount := 0
	for {
		secret, err := cli.GetSensitiveUserInput("Enter secret: ", printer)
		if err != nil {
			return "", fmt.Errorf("unable to get sercret: %w", err)
		}

		decrypted, err := password.GetDecrypted(secret, keyLen)
		if err != nil {
			if tryCount >= maxRetries {
				return "", fmt.Errorf("unable to check secret: %w", err)
			}

			printer.Warning("Incorrect secret, try again")
			tryCount++
		} else {
			return decrypted, nil
		}
	}
}

// getPasswordGetterByGenerateAndLengthFlag return function for getting password based on flags -g and --length
func getPasswordGetterByGenerateAndLengthFlag(
	cmd *cobra.Command,
	genFlag, lenFlag, passwordAlias string,
	printer Printer,
	conf *config.Config,
) (func() (string, error), error) {
	needGenerate, err := cmd.Flags().GetBool(genFlag)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s flag: %w", genFlag, err)
	}

	if needGenerate {
		length, err := cmd.Flags().GetInt(lenFlag)
		if err != nil {
			return nil, fmt.Errorf("unable to get %s flag: %w", lenFlag, err)
		}

		return func() (string, error) {
			userPassword, err := getGeneratedPassword(length, conf, printer)
			if err != nil {
				return "", fmt.Errorf("unable to get generated password: %w", err)
			}
			return userPassword, nil
		}, nil
	} else {
		return func() (string, error) {
			return getSecretWithConfirmation(passwordAlias, "Passwords are not equal", printer), nil
		}, nil
	}
}
