# goseal

goseal is a simple CLI tool to easily create kubernetes secrets using kubectl and kubeseal (optional).

## Prerequisites

[kubectl](https://kubernetes.io/docs/reference/kubectl/kubectl/)

[kubeseal](https://fluxcd.io/docs/guides/sealed-secrets/)

## Installation

```sh
go get -u github.com/MaxBreida/goseal
go install github.com/MaxBreida/goseal
```

## Usage

### Available commands

```text
NAME:
   goseal - Used to automatically generate kubernetes secret files (and optionally seal them)

USAGE:
   goseal [global options] command [command options] [arguments...]

COMMANDS:
   yaml, y  Creates a secret file from yaml input
   file     Creates a secret file from file input
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

### Create secret from yaml key-value pairs

```text
NAME:
   yaml - Create a secret file with key-value pairs as in the yaml file

USAGE:
   yaml [command options] [arguments...]

DESCRIPTION:
   creates a sealed secret from yaml input key-value pairs

OPTIONS:
   --help, -h                      show help (default: false)
   --namespace value, --nsp value  the namespace of the secret
   --file value, -f value          the input file in yaml format
   --name value, -n value          the secret name
   --cert value, -c value          if set, will run kubeseal with given cert
```

To create an unsealed kubernetes, run:

```sh
goseal yaml -f my-file.yaml -nsp my-namespace -n my-secret > output.yaml
```

To seal the secret, simply add the --cert or -c flag to the command:

```sh
goseal yaml -f my-file.yaml -nsp my-namespace -n my-secret -c path/to/my/cert.pem > output.yaml
```
