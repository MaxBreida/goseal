# goseal

goseal is a simple CLI tool to easily create kubernetes secrets using kubectl and kubeseal (optional).

## Prerequisites

[kubectl](https://kubernetes.io/docs/reference/kubectl/kubectl/)

[kubeseal](https://fluxcd.io/docs/guides/sealed-secrets/)

## Installation

```sh
go get -u github.com/MaxBreida/goseal
go install github.com/MaxBreida/goseal@latest
```

## Usage

### Available commands

```text
NAME:
   goseal - Used to automatically generate kubernetes secret files (and optionally seal them)

USAGE:
   goseal [global options] command [command options] [arguments...]

VERSION:
   vx.x.x

COMMANDS:
   yaml, y  Create a secret file with key-value pairs as in the yaml file
   file     Create a secret with a file as secret value.
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
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
   --help, -h                     show help (default: false)
   --namespace value, -n value    the namespace of the secret
   --file value, -f value         the input file in yaml format
   --secret-name value, -s value  the secret name
   --cert value, -c value         if set, will run kubeseal with given cert
```

To create an unsealed secret, run:

```sh
goseal yaml -f my-file.yaml -n my-namespace -s my-secret > output.yaml
```

To seal the secret, simply add the --cert or -c flag to the command:

```sh
goseal yaml -f my-file.yaml -n my-namespace -s my-secret -c path/to/my/cert.pem > output.yaml
```

### Create secret from any file

This command lets you create a secret from any given file. For this, a secret-key is required under
which the file can be accessed.

```text
NAME:
   file - Create a secret with a file as secret value.

USAGE:
   file [command options] [arguments...]

DESCRIPTION:
   creates a (sealed) kubernetes secret with a file as secret value

OPTIONS:
   --help, -h                     show help (default: false)
   --namespace value, -n value    the namespace of the secret
   --file value, -f value         the input file in yaml format
   --secret-name value, -s value  the secret name
   --cert value, -c value         if set, will run kubeseal with given cert
   --key value, -k value          the secret key, under which the file can be accessed
```

To create an unsealed secret, run:

```sh
goseal file -f my-file -n my-namespace -s my-secret -k my-key > output.yaml
```

To seal the secret, simply add the --cert or -c flag to the command:

```sh
goseal file -f my-file -n my-namespace -s my-secret -k my-key -c path/to/my/cert.pem > output.yaml
```
