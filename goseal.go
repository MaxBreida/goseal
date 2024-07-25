package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

func main() {
	app := &cli.App{
		Name:    "goseal",
		Usage:   "Used to automatically generate kubernetes secret files (and optionally seal them)",
		Version: "v0.3.0",
		Commands: []*cli.Command{
			{
				Name:        "yaml",
				HelpName:    "yaml",
				Description: "creates a sealed secret from yaml input key-value pairs",
				Usage:       "Create a secret file with key-value pairs as in the yaml file",
				Aliases:     []string{"y"},
				Flags:       getStandardFlags(),
				Action:      Yaml,
			},
			{
				Name:        "file",
				HelpName:    "file",
				Description: "creates a (sealed) kubernetes secret with a file as secret value",
				Usage:       "Create a secret with a file as secret value.",
				Action:      File,
				Flags: append(getStandardFlags(), &cli.StringFlag{
					Name:     "key",
					Usage:    "the secret key, under which the file can be accessed",
					Aliases:  []string{"k"},
					Required: true,
				}),
			},
		},
	}

	app.EnableBashCompletion = true

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getStandardFlags() []cli.Flag {
	return []cli.Flag{
		cli.BashCompletionFlag,
		cli.HelpFlag,
		&cli.StringFlag{
			Name:     "namespace",
			Usage:    "the namespace of the secret",
			Required: true,
			Aliases:  []string{"n"},
		},
		&cli.StringFlag{
			Name:     "file",
			Usage:    "the input file in yaml format",
			Required: true,
			Aliases:  []string{"f"},
		},
		&cli.StringFlag{
			Name:     "secret-name",
			Usage:    "the secret name",
			Required: true,
			Aliases:  []string{"s"},
		},
		&cli.StringFlag{
			Name:    "cert",
			Usage:   "if set, will run kubeseal with given cert",
			Aliases: []string{"c"},
		},
	}
}

// ErrEmptyFile is returned if the provided file has no content.
var ErrEmptyFile = errors.New("file content is empty")

// Yaml is a cli command
func Yaml(c *cli.Context) error {
	filePath := c.String("file")
	namespace := c.String("namespace")
	secretName := c.String("secret-name")
	certPath := c.String("cert")

	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if len(file) == 0 {
		return ErrEmptyFile
	}

	var secrets map[string]string

	if err := yaml.Unmarshal(file, &secrets); err != nil {
		return err
	}

	if certPath != "" {
		return sealSecret(secrets, secretName, namespace, certPath)
	}

	return createSecret(secrets, secretName, namespace)
}

// File is a cli command
func File(c *cli.Context) error {
	filePath := c.String("file")
	secretKey := c.String("key")
	namespace := c.String("namespace")
	secretName := c.String("secret-name")
	certPath := c.String("cert")

	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if len(file) == 0 {
		return ErrEmptyFile
	}

	secrets := map[string]string{secretKey: string(file)}

	if certPath != "" {
		return sealSecret(secrets, secretName, namespace, certPath)
	}

	return createSecret(secrets, secretName, namespace)
}

// regexCreationTimestamp is a regex used to remove the creationTimestamp from the output of kubectl create secret.
var regexCreationTimestamp = regexp.MustCompile(`\s*creationTimestamp: null`)

// runs the kubectl create secret command and prints the output to stdout.
func createSecret(secrets map[string]string, secretName, namespace string) error {
	kubectlCreateSecret := getCreateSecretFileCmd(secrets, secretName, namespace)

	var stdout bytes.Buffer
	kubectlCreateSecret.Stdout = &stdout

	if err := runCommand(kubectlCreateSecret); err != nil {
		return err
	}

	outputWithCreationTimestamp := stdout.String()
	output := regexCreationTimestamp.ReplaceAllString(outputWithCreationTimestamp, "")

	fmt.Println(output)

	return nil
}

// runs the kubectl create secret command, pipes the output to the kubeseal command and prints the output to stdout.
func sealSecret(secrets map[string]string, secretName, namespace, certPath string) error {
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
		return err
	}

	if err := kubeseal.Start(); err != nil {
		return err
	}

	if err = runCommand(kubectlCreateSecret); err != nil {
		return err
	}

	if err = kubeseal.Wait(); err != nil {
		return errors.New(getErrText(err, kubeseal.Args, stderr.String()))
	}

	outputWithCreationTimestamp := stdout.String()
	output := regexCreationTimestamp.ReplaceAllString(outputWithCreationTimestamp, "")

	fmt.Println(output)

	return nil
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
