package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

// GetSecretsFromFile reads a filepath and returns a map of secretKey:secretValue pairs.
func GetSecretsFromFile(fileMode, filePath, secretKey string) (map[string]string, error) {
	var secrets map[string]string

	switch fileMode {
	case fileModeFile:
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		if len(file) == 0 {
			return nil, ErrEmptyFile
		}

		secrets = map[string]string{secretKey: string(file)}
	case "yaml":
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		if len(file) == 0 {
			return nil, ErrEmptyFile
		}

		var secrets map[string]string

		if err := yaml.Unmarshal(file, &secrets); err != nil {
			return nil, err
		}
	}

	return secrets, nil
}

// CreateSecret runs the kubectl create secret command and returns the file content.
func CreateSecret(secrets map[string]string, secretName, namespace string) ([]byte, error) {
	kubectlCreateSecret := getCreateSecretFileCmd(secrets, secretName, namespace)

	var stdout bytes.Buffer
	kubectlCreateSecret.Stdout = &stdout

	if err := runCommand(kubectlCreateSecret); err != nil {
		return nil, err
	}

	return stdout.Bytes(), nil
}

// SealSecret runs the kubectl create secret command, pipes the output to the kubeseal command and returns the file content.
func SealSecret(secrets map[string]string, secretName, namespace, certPath string) ([]byte, error) {
	kubectlCreateSecret := getCreateSecretFileCmd(secrets, secretName, namespace)
	kubeseal := exec.Command("kubeseal", "--format", "yaml", "--cert", certPath)

	var (
		err            error
		stdout, stderr bytes.Buffer
	)

	kubeseal.Stdout = &stdout
	kubeseal.Stderr = &stderr

	// Get stdout of first command and attach it to stdin of second command.
	kubeseal.Stdin, err = kubectlCreateSecret.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := kubeseal.Start(); err != nil {
		return nil, err
	}

	if err = runCommand(kubectlCreateSecret); err != nil {
		return nil, err
	}

	if err = kubeseal.Wait(); err != nil {
		return nil, errors.New(getErrText(err, kubeseal.Args, stderr.String()))
	}

	return stdout.Bytes(), nil
}

// retrieve a printable error text from cmd errors
func getErrText(err error, cmdArgs []string, stdErr string) string {
	text := fmt.Sprintf(
		"command '%s' failed: %s",
		strings.Join(cmdArgs, " "),
		err.Error(),
	)

	errText := strings.TrimSpace(stdErr)
	if len(errText) > 0 {
		text += "\n" + errText
	}

	return text
}

func runCommand(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(getErrText(err, cmd.Args, stderr.String()))
	}

	return nil
}

// creates
func getCreateSecretFileCmd(secrets map[string]string, secretName, namespace string) *exec.Cmd {
	args := []string{
		"create",
		"secret",
		"generic",
		secretName,
		"-n",
		namespace,
		"--dry-run",
		"-o",
		"yaml",
	}

	for k, v := range secrets {
		args = append(args, fmt.Sprintf("--from-literal=%s=%s", k, v))
	}

	return exec.Command("kubectl", args...)
}
