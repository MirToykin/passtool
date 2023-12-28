# Passtool: Secure Password Management CLI

## Overview
Passtool is a command-line tool for securely managing your passwords. It allows users to securely store, retrieve, 
and manage passwords for various services and accounts. The tool uses strong encryption to ensure your passwords 
are safe and supports environment variables for flexible configuration.

### Features
- Securely add, set, and retrieve passwords.
- Password encryption and decryption using a secret key.
- Support for environment variables for easy configuration.
- Automated database backups.
- Ability to generate strong passwords.

## Getting Started

### Prerequisites
- Ensure you have a suitable environment to run Go applications.

## Configuration
Set the following required and optional environment variables:

### Required Variables:
- `PASSTOOL_STORAGE_PATH`: Path to the directory where data will be stored (e.g., `/Users/me/passtool`). Data is kept encrypted in this location.

### Optional Variables:
- `PASSTOOL_BACKUP_INDEX`: Perform a DB backup for each N added passwords. Default is 5.
- `PASSTOOL_BACKUP_COUNT`: Number of backups to retain. Default is 5.
- `PASSTOOL_DEFAULT_PASSWORD_LENGTH`: Default length for generated passwords. Default is 12.

## Usage

### Commands:
1. `passtool add`: Add a new password.
    - `-g, --generate`: Generate a secure password.
    - `--length int`: Specify the length of the generated password (default 12).

2. `passtool set`: Set a new password for an existing account of a service.
    - `-g, --generate`: Generate a secure password.
    - `--length int`: Specify the length of the generated password (default 12).

3. `passtool get`: Retrieve a password for a specific account of a service.

4. `passtool del`: Delete an account and its associated password for a service.

5. `passtool list`: Print the list of available services.
    - `-a, --accounts`: Print related accounts along with each service.

6. `passtool change-secret`: Change the secret key used for encryption.

7. `passtool requirements`: Print requirements for the service to work.

## Security
- Passtool does not store your secret key; it must be provided each time for encryption and decryption.
- Ensure you keep your secret key safe and never share it.

## Contact
miroslavtoikin@gmail.com

