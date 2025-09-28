package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

type Credential struct {
	Type     string `json:"type"`     // "aws" or "git"
	Name     string `json:"name"`     // Display name like "aws-project1", "github-self"
	Username string `json:"username"` // For git: username, for AWS: access key ID
	Password string `json:"password"` // For git: password/token, for AWS: secret key
	Extra    string `json:"extra"`    // For AWS: region, for git: could be empty
}

type CredentialStore struct {
	filePath string
}

// Get secure input (password/PIN) without echoing
func getSecureInput(prompt string) string {
	fmt.Print(prompt)
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println("\nError reading input:", err)
		return ""
	}
	fmt.Println() // New line after password input
	return string(bytePassword)
}

// Get regular input
func getInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// Encrypt data using AES-GCM
func encrypt(plainText, pin string) (string, error) {
	key := sha256.Sum256([]byte(pin))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(cipherText), nil
}

// Decrypt data using AES-GCM
func decrypt(cipherHex, pin string) (string, error) {
	key := sha256.Sum256([]byte(pin))
	data, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("invalid ciphertext")
	}

	nonce, cipherText := data[:nonceSize], data[nonceSize:]
	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

// Get credentials file path based on OS
func getCredentialsPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".credentials.enc"
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local", "credentials.enc")
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "credentials.enc")
	default: // linux and others
		return filepath.Join(homeDir, ".config", "credentials.enc")
	}
}

// Initialize credential store
func NewCredentialStore() *CredentialStore {
	return &CredentialStore{
		filePath: getCredentialsPath(),
	}
}

// Load credentials from encrypted file
func (cs *CredentialStore) loadCredentials(pin string) ([]Credential, error) {
	if _, err := os.Stat(cs.filePath); os.IsNotExist(err) {
		return []Credential{}, nil
	}

	encData, err := os.ReadFile(cs.filePath)
	if err != nil {
		return nil, err
	}

	jsonData, err := decrypt(string(encData), pin)
	if err != nil {
		return nil, fmt.Errorf("wrong PIN or corrupted data")
	}

	var credentials []Credential
	err = json.Unmarshal([]byte(jsonData), &credentials)
	return credentials, err
}

// Save credentials to encrypted file
func (cs *CredentialStore) saveCredentials(credentials []Credential, pin string) error {
	// Ensure directory exists
	dir := filepath.Dir(cs.filePath)
	os.MkdirAll(dir, 0700)

	jsonData, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return err
	}

	encData, err := encrypt(string(jsonData), pin)
	if err != nil {
		return err
	}

	return os.WriteFile(cs.filePath, []byte(encData), 0600)
}

// Add new credential
func (cs *CredentialStore) addCredential() {
	fmt.Println("\n=== Add New Credential ===")

	credType := getInput("Type (aws/git): ")
	if credType != "aws" && credType != "git" {
		fmt.Println("Invalid type. Must be 'aws' or 'git'")
		return
	}

	name := getInput("Name (e.g., aws-project1, github-self): ")
	if name == "" {
		fmt.Println("Name cannot be empty")
		return
	}

	var username, password, extra string

	if credType == "aws" {
		username = getInput("AWS Access Key ID: ")
		password = getSecureInput("AWS Secret Access Key: ")
		extra = getInput("AWS Region (optional): ")
	} else {
		username = getInput("Git Username: ")
		password = getSecureInput("Git Password/Token: ")
	}

	pin := getSecureInput("Enter PIN to encrypt credentials: ")
	if pin == "" {
		fmt.Println("PIN cannot be empty")
		return
	}

	// Load existing credentials
	credentials, err := cs.loadCredentials(pin)
	if err != nil && !strings.Contains(err.Error(), "no such file") {
		fmt.Println("Error loading existing credentials:", err)
		return
	}

	// Check for duplicate names
	for _, cred := range credentials {
		if cred.Name == name {
			fmt.Printf("Credential with name '%s' already exists\n", name)
			return
		}
	}

	// Add new credential
	newCred := Credential{
		Type:     credType,
		Name:     name,
		Username: username,
		Password: password,
		Extra:    extra,
	}

	credentials = append(credentials, newCred)

	// Save credentials
	if err := cs.saveCredentials(credentials, pin); err != nil {
		fmt.Println("Error saving credentials:", err)
		return
	}

	fmt.Println("Credential added successfully!")
}

// List all credentials
func (cs *CredentialStore) listCredentials() {
	pin := getSecureInput("Enter PIN to access credentials: ")
	if pin == "" {
		fmt.Println("PIN cannot be empty")
		return
	}

	credentials, err := cs.loadCredentials(pin)
	if err != nil {
		fmt.Println("Error loading credentials:", err)
		return
	}

	if len(credentials) == 0 {
		fmt.Println("No credentials found")
		return
	}

	fmt.Println("\n=== Available Credentials ===")
	for i, cred := range credentials {
		fmt.Printf("%d. %s (%s)\n", i+1, cred.Name, cred.Type)
	}

	choice := getInput("\nEnter number to login (or 'q' to quit): ")
	if choice == "q" || choice == "" {
		return
	}

	var selectedIndex int
	if _, err := fmt.Sscanf(choice, "%d", &selectedIndex); err != nil || selectedIndex < 1 || selectedIndex > len(credentials) {
		fmt.Println("Invalid selection")
		return
	}

	cs.performLogin(credentials[selectedIndex-1])
}

// Perform login based on credential type
func (cs *CredentialStore) performLogin(cred Credential) {
	fmt.Printf("\nLogging in with %s...\n", cred.Name)

	if cred.Type == "aws" {
		cs.performAWSLogin(cred)
	} else if cred.Type == "git" {
		cs.performGitLogin(cred)
	}
}

// Perform AWS login
func (cs *CredentialStore) performAWSLogin(cred Credential) {
	fmt.Println("AWS Login Options:")
	fmt.Println("1. Configure AWS CLI profile")
	fmt.Println("2. Export environment variables (script)")
	fmt.Println("3. Launch shell with credentials")
	fmt.Println("4. Test connection only")

	choice := getInput("Choose option (1-4): ")

	switch choice {
	case "1":
		cs.configureAWSProfile(cred)
	case "2":
		cs.exportAWSEnvScript(cred)
	case "3":
		cs.launchAWSShell(cred)
	case "4":
		cs.testAWSConnection(cred)
	default:
		fmt.Println("Invalid option")
	}
}

// Configure AWS CLI profile
func (cs *CredentialStore) configureAWSProfile(cred Credential) {
	profileName := getInput("Enter AWS profile name (default: " + cred.Name + "): ")
	if profileName == "" {
		profileName = cred.Name
	}

	// Configure AWS profile
	cmd1 := exec.Command("aws", "configure", "set", "aws_access_key_id", cred.Username, "--profile", profileName)
	if err := cmd1.Run(); err != nil {
		fmt.Printf("Failed to set access key: %v\n", err)
		return
	}

	cmd2 := exec.Command("aws", "configure", "set", "aws_secret_access_key", cred.Password, "--profile", profileName)
	if err := cmd2.Run(); err != nil {
		fmt.Printf("Failed to set secret key: %v\n", err)
		return
	}

	if cred.Extra != "" {
		cmd3 := exec.Command("aws", "configure", "set", "region", cred.Extra, "--profile", profileName)
		if err := cmd3.Run(); err != nil {
			fmt.Printf("Failed to set region: %v\n", err)
			return
		}
	}

	fmt.Printf("AWS profile '%s' configured successfully!\n", profileName)
	fmt.Printf("Use with: aws --profile %s <command>\n", profileName)
	fmt.Printf("Or set as default: export AWS_PROFILE=%s\n", profileName)

	// Test the profile
	fmt.Println("\nTesting profile...")
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--profile", profileName)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Profile test failed: %v\n", err)
	} else {
		fmt.Println("Profile test successful!")
		fmt.Printf("Identity: %s", string(output))
	}
}

// Export environment variables script
func (cs *CredentialStore) exportAWSEnvScript(cred Credential) {
	scriptContent := ""
	fileName := ""

	if runtime.GOOS == "windows" {
		fileName = "aws_env.bat"
		scriptContent = fmt.Sprintf(`@echo off
set AWS_ACCESS_KEY_ID=%s
set AWS_SECRET_ACCESS_KEY=%s`, cred.Username, cred.Password)
		if cred.Extra != "" {
			scriptContent += fmt.Sprintf(`
set AWS_DEFAULT_REGION=%s`, cred.Extra)
		}
		scriptContent += `
echo AWS credentials set in environment
echo Run this script or copy these commands to your terminal
pause`
	} else {
		fileName = "aws_env.sh"
		scriptContent = fmt.Sprintf(`#!/bin/bash
export AWS_ACCESS_KEY_ID="%s"
export AWS_SECRET_ACCESS_KEY="%s"`, cred.Username, cred.Password)
		if cred.Extra != "" {
			scriptContent += fmt.Sprintf(`
export AWS_DEFAULT_REGION="%s"`, cred.Extra)
		}
		scriptContent += `
echo "AWS credentials exported to environment"
echo "Run: source aws_env.sh"
echo "Or copy-paste the export commands above"`
	}

	if err := os.WriteFile(fileName, []byte(scriptContent), 0600); err != nil {
		fmt.Printf("Failed to create script: %v\n", err)
		return
	}

	fmt.Printf("Environment script created: %s\n", fileName)
	if runtime.GOOS == "windows" {
		fmt.Println("Run: aws_env.bat")
	} else {
		fmt.Println("Run: source aws_env.sh")
		// Make it executable
		os.Chmod(fileName, 0700)
	}
}

// Launch shell with AWS credentials
func (cs *CredentialStore) launchAWSShell(cred Credential) {
	fmt.Println("Launching new shell with AWS credentials...")

	var cmd *exec.Cmd
	env := os.Environ()
	env = append(env, "AWS_ACCESS_KEY_ID="+cred.Username)
	env = append(env, "AWS_SECRET_ACCESS_KEY="+cred.Password)
	if cred.Extra != "" {
		env = append(env, "AWS_DEFAULT_REGION="+cred.Extra)
	}

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd")
	} else {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}
		cmd = exec.Command(shell)
	}

	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("AWS credentials are available in this shell session")
	fmt.Println("Type 'exit' to return to credential manager")

	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to launch shell: %v\n", err)
	}
}

// Test AWS connection
func (cs *CredentialStore) testAWSConnection(cred Credential) {
	fmt.Println("Testing AWS connection...")

	cmd := exec.Command("aws", "sts", "get-caller-identity")
	cmd.Env = append(os.Environ(),
		"AWS_ACCESS_KEY_ID="+cred.Username,
		"AWS_SECRET_ACCESS_KEY="+cred.Password,
	)
	if cred.Extra != "" {
		cmd.Env = append(cmd.Env, "AWS_DEFAULT_REGION="+cred.Extra)
	}

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("AWS connection test failed: %v\n", err)
		if strings.Contains(err.Error(), "executable file not found") {
			fmt.Println("Make sure AWS CLI is installed and in your PATH")
		}
	} else {
		fmt.Println("AWS connection successful!")
		fmt.Printf("Identity: %s", string(output))
	}
}

// Perform Git login
func (cs *CredentialStore) performGitLogin(cred Credential) {
	// For GitHub/GitLab, we can configure git credentials
	fmt.Println("Available Git login options:")
	fmt.Println("1. Set global git credentials")
	fmt.Println("2. Login to GitHub CLI (if installed)")

	choice := getInput("Choose option (1-2): ")

	switch choice {
	case "1":
		// Set git credentials globally
		cmd1 := exec.Command("git", "config", "--global", "user.name", cred.Username)
		if err := cmd1.Run(); err != nil {
			fmt.Printf("Failed to set git username: %v\n", err)
			return
		}

		fmt.Println("Git username configured globally")
		fmt.Println("Note: For HTTPS authentication, use the token as password when prompted")
		fmt.Printf("Your token: %s\n", cred.Password)

	case "2":
		// GitHub CLI login
		fmt.Println("Attempting GitHub CLI login...")
		cmd := exec.Command("gh", "auth", "login", "--with-token")
		cmd.Stdin = strings.NewReader(cred.Password)

		if err := cmd.Run(); err != nil {
			fmt.Printf("GitHub CLI login failed (make sure 'gh' is installed): %v\n", err)
			fmt.Println("You can manually run: echo 'YOUR_TOKEN' | gh auth login --with-token")
		} else {
			fmt.Println("GitHub CLI login successful!")
		}

	default:
		fmt.Println("Invalid option")
	}
}

// Delete credential
func (cs *CredentialStore) deleteCredential() {
	pin := getSecureInput("Enter PIN to access credentials: ")
	if pin == "" {
		fmt.Println("PIN cannot be empty")
		return
	}

	credentials, err := cs.loadCredentials(pin)
	if err != nil {
		fmt.Println("Error loading credentials:", err)
		return
	}

	if len(credentials) == 0 {
		fmt.Println("No credentials found")
		return
	}

	fmt.Println("\n=== Delete Credential ===")
	for i, cred := range credentials {
		fmt.Printf("%d. %s (%s)\n", i+1, cred.Name, cred.Type)
	}

	choice := getInput("\nEnter number to delete (or 'q' to quit): ")
	if choice == "q" || choice == "" {
		return
	}

	var selectedIndex int
	if _, err := fmt.Sscanf(choice, "%d", &selectedIndex); err != nil || selectedIndex < 1 || selectedIndex > len(credentials) {
		fmt.Println("Invalid selection")
		return
	}

	// Confirm deletion
	toDelete := credentials[selectedIndex-1]
	confirm := getInput(fmt.Sprintf("Are you sure you want to delete '%s'? (y/N): ", toDelete.Name))
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println("Deletion cancelled")
		return
	}

	// Remove credential
	credentials = append(credentials[:selectedIndex-1], credentials[selectedIndex:]...)

	// Save updated credentials
	if err := cs.saveCredentials(credentials, pin); err != nil {
		fmt.Println("Error saving credentials:", err)
		return
	}

	fmt.Printf("Credential '%s' deleted successfully!\n", toDelete.Name)
}

func main() {
	store := NewCredentialStore()

	for {
		fmt.Println("\n=== Credential Manager ===")
		fmt.Println("1. Add credential")
		fmt.Println("2. List & Login")
		fmt.Println("3. Delete credential")
		fmt.Println("4. Exit")

		choice := getInput("\nChoose option (1-4): ")

		switch choice {
		case "1":
			store.addCredential()
		case "2":
			store.listCredentials()
		case "3":
			store.deleteCredential()
		case "4":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid option. Please choose 1-4.")
		}
	}
}
