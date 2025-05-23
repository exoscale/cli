package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

var (
	errNoSSHCommand = errors.New("ssh command cannot be empty")
	errNoUser       = errors.New("ssh user cannot be empty")
	errNoSSHKey     = errors.New("no valid SSH private keys found")
)

func RunCmd(publicIP net.IP, user, port, sshKeyName, cmd string) (string, error) {
	// Check if the user is empty
	if user == "" {
		return "", errNoUser
	}

	// Make sure the command is not empty
	if cmd == "" {
		return "", errNoSSHCommand
	}

	if port == "" {
		port = "22" // Default SSH port
	}

	signers, err := loadSSHKeys(sshKeyName)
	if err != nil {
		return "", err
	}

	// Define SSH server and credentials
	server := fmt.Sprintf("%s:%s", publicIP.String(), port)

	// Create SSH client configuration
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to the SSH server
	client, err := ssh.Dial("tcp", server, config)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close SSH client: %v", err)
		}
	}()

	// Create a new session
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer func() {
		if err := session.Close(); err != nil && err != io.EOF {
			log.Printf("Failed to close SSH session: %v", err)
		}
	}()

	// Run the command and capture the output
	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	if err := session.Start(cmd); err != nil {
		return "", fmt.Errorf("failed to start command: %v", err)
	}

	if err := session.Wait(); err != nil {
		return "", fmt.Errorf("failed to wait for command: %v", err)
	}

	return stdoutBuf.String(), nil

}

func loadSSHKeys(sshKeyName string) ([]ssh.Signer, error) {
	// Check for the existence of RSA and ed25519 keys
	rsaKeyPath := filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	ed25519KeyPath := filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519")

	var signers []ssh.Signer

	if sshKeyName != "" {
		customSSHKeyPath := filepath.Join(os.Getenv("HOME"), ".ssh", sshKeyName)
		if key, err := os.ReadFile(customSSHKeyPath); err == nil {
			if signer, err := ssh.ParsePrivateKey(key); err == nil {
				signers = append(signers, signer)
			}
		}
	}

	if key, err := os.ReadFile(rsaKeyPath); err == nil {
		if signer, err := ssh.ParsePrivateKey(key); err == nil {
			signers = append(signers, signer)
		}
	}

	if key, err := os.ReadFile(ed25519KeyPath); err == nil {
		if signer, err := ssh.ParsePrivateKey(key); err == nil {
			signers = append(signers, signer)
		}
	}

	if len(signers) == 0 {
		return signers, errNoSSHKey
	}

	return signers, nil
}
