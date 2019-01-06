# Drycc Command-Line Interface

drycc-cli is the command-line client for the [controller](/controller). It provides
access to many functions related to deploying and managing applications.

## Installation

Pre-built binaries are available for Mac OS X, Linux, and Windows. Once
installed, these binaries will automatically update themselves when new releases
are available.

To install the latest release on OS X or Linux, run this command in a terminal:

```text
L=/usr/local/bin/drycc && curl -sSL -A "`uname -sp`" https://dl.drycc.cc/cli | zcat >$L && chmod +x $L
```

To install the latest release on Windows, run this command in PowerShell:

```text
(New-Object Net.WebClient).DownloadString('https://dl.drycc.cc/cli.ps1') | iex
```


## Usage

The basic usage is:

```text
drycc [-a app] <command> [options] [arguments]
```

For a list of commands and usage instructions, run `drycc help`.

## Credits

drycc-cli is a fork of Heroku's [hk](https://github.com/heroku/hk).
