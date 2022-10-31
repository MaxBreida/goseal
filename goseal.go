package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "goseal",
		Usage:   "Used to automatically generate kubernetes secret files (and optionally seal them)",
		Version: "v0.2.0",
		Commands: []*cli.Command{
			{
				Name:        "ui",
				Description: "starts the goseal TUI",
				Action:      StartUI,
			},
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

	secrets, err := GetSecretsFromFile("yaml", filePath, "")
	if err != nil {
		return err
	}

	var result []byte
	if certPath != "" {
		result, err = SealSecret(secrets, secretName, namespace, certPath)
	} else {
		result, err = CreateSecret(secrets, secretName, namespace)
	}
	if err != nil {
		return err
	}

	fmt.Println(string(result))

	return nil
}

// File is a cli command
func File(c *cli.Context) error {
	filePath := c.String("file")
	secretKey := c.String("key")
	namespace := c.String("namespace")
	secretName := c.String("secret-name")
	certPath := c.String("cert")

	secrets, err := GetSecretsFromFile("yaml", filePath, secretKey)
	if err != nil {
		return err
	}

	var result []byte
	if certPath != "" {
		result, err = SealSecret(secrets, secretName, namespace, certPath)
	} else {
		result, err = CreateSecret(secrets, secretName, namespace)
	}
	if err != nil {
		return err
	}

	fmt.Println(string(result))

	return nil
}
