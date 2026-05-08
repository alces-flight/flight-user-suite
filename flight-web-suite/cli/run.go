package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

func RunLocal[T any](ctx context.Context, description string, tool Tool, username string, args ...string) (T, error) {
	var zero T

	cmd, err := BuildLocalCommand(ctx, tool, username, args...)
	if err != nil {
		return zero, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	var response T
	if decodeErr := json.Unmarshal(stdout.Bytes(), &response); decodeErr == nil {
		return response, nil
	}

	if runErr != nil {
		if stderr.Len() != 0 {
			return zero, fmt.Errorf("%s (local): %s", description, stderr.String())
		}
		return zero, fmt.Errorf("%s (local): %w", description, runErr)
	}
	return zero, fmt.Errorf("decoding response (local): %s", stdout.String())
}

func RunRemote[T any](description string, tool RemoteTool, username, remoteHost string, args []string) (T, error) {
	var zero T

	sess, cleanup, err := remoteSession(username, remoteHost, tool.RemoteConfig())
	defer cleanup()
	if err != nil {
		return zero, fmt.Errorf("establishing remote session: %w", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	sess.Stdout = &stdout
	sess.Stderr = &stderr
	cmd := strings.Join(append([]string{tool.ToolPath()}, args...), " ")
	runErr := sess.Run(cmd)

	var response T
	if decodeErr := json.Unmarshal(stdout.Bytes(), &response); decodeErr == nil {
		return response, nil
	}

	if runErr != nil {
		if stderr.Len() != 0 {
			return zero, fmt.Errorf("%s (remote): %s", description, stderr.String())
		}
		return zero, fmt.Errorf("%s (remote): %w", description, runErr)
	}
	return zero, fmt.Errorf("decoding response (remote): %s", stdout.String())
}

func remoteSession(username, remoteHost string, config RemoteConfig) (*ssh.Session, func(), error) {
	var conn *ssh.Client
	var sess *ssh.Session

	cleanup := func() {
		if sess != nil {
			_ = sess.Close()
		}
		if conn != nil {
			_ = conn.Close()
		}
	}

	homedir := filepath.Clean(fmt.Sprintf("%s/%s", config.HomeDirFallback, username))
	user, err := user.Lookup(username)
	if err == nil && user != nil {
		homedir = user.HomeDir
	}
	publicKey, err := publicKey(filepath.Join(homedir, ".ssh", config.SshKeyName))
	if err != nil {
		return nil, cleanup, err
	}

	clientConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{publicKey},

		// TODO: Do we need to change this? We create an SSH keypair via the
		// ssh-keypair-generation hook and configure the ssh config with
		// `StrictHostKeyChecking no`. Would changing this just be a pretence?
		//
		// hostKeyCallback, err := knownhosts.New(filepath.Join(homedir, ".ssh", "known_hosts"))
		// if err != nil {
		// 	return nil, cleanup, fmt.Errorf("creating hostKeyCallback: %w", err)
		// }
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         config.ConnectionTimeout,
	}
	conn, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", remoteHost), clientConfig)
	if err != nil {
		return nil, cleanup, err
	}

	sess, err = conn.NewSession()
	if err != nil {
		return nil, cleanup, err
	}

	return sess, cleanup, nil
}

func publicKey(path string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}
	return ssh.PublicKeys(signer), nil
}
