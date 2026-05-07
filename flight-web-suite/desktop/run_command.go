package desktop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"golang.org/x/crypto/ssh"
)

func RunLocal[T any](description string, cmd *exec.Cmd) (T, error) {
	var zero T

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

func runRemoteCommand[T any](description, username, cmd, remoteHost string) (T, error) {
	var zero T

	sess, cleanup, err := remoteSession(username, remoteHost)
	defer cleanup()
	if err != nil {
		return zero, fmt.Errorf("establishing remote session: %w", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	sess.Stdout = &stdout
	sess.Stderr = &stderr
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

func remoteSession(username, remoteHost string) (*ssh.Session, func(), error) {
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

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			// TODO: Take this from configuration.
			// Need to get user's home directory.  Attempt to use
			// `os/user.Current`? Fallback to configurable `/home/<username>`?
			// Add documentation that password-less SSH is required. Mention ssh-keypair-generation hook.
			publicKey("/home/ben/.ssh/id_flightcluster"),
		},
		// TODO: Replace with known_hosts parsing.
		// "golang.org/x/crypto/ssh/knownhosts"
		// kh.New("/Users/user/.ssh/known_hosts")
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// TODO: Take this from configuration.
		Timeout: time.Duration(5 * time.Second),
	}
	var err error
	conn, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", remoteHost), config)
	if err != nil {
		return nil, cleanup, err
	}

	sess, err = conn.NewSession()
	if err != nil {
		return nil, cleanup, err
	}

	return sess, cleanup, nil
}

func publicKey(path string) ssh.AuthMethod {
	key, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}
	return ssh.PublicKeys(signer)
}
