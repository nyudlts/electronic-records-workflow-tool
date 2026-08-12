# electronic-records-workflow-tool v1.2.0

## Build

```bash
go build -o ewt main.go
```

## Commands

```text
Available Commands:
  amatica     ewt Archivematica commands
  aspace      ewt ArchivesSpace commands
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  project     ewt project commands
  rstar       ewt AIP commands
  sip         ewt SIP commands
  source      ewt source commands
  version     print the version of ewt
```

## `amatica`

Archivematica-related commands.

### Subcommands

* `prep`
* `transfer`

## `aspace`

ArchivesSpace-related commands.

### Subcommands

* `check`

## `completion`

Generates the autocompletion script for the specified shell.

## `help`

Displays help information for `ewt` and its commands.

## `project`

Project initialization and closeout commands.

### `init`

Initializes a new project.

```text
ewt project init -c <collection-code> -s <source-path>
```

For example:

```bash
ewt project init -c dlts_test -s /mnt/amatica/testing/ewt-test-summer-2026/to_rstar/
```

`project init` creates a new project directory in the current location. The project directory contains the directories required to run `ewt`, along with a `config.json` file containing the configuration required by the application.

For example:

```json
{
  "sip-location": "/home/archivematica/ewt/dlts_test/sip",
  "source-location": "/mnt/amatica/testing/ewt-test-summer-2026/to_rstar/",
  "partner-code": "dlts",
  "collection-code": "dlts_test",
  "project-location": "/home/archivematica/ewt/dlts_test",
  "log-location": "/home/archivematica/ewt/dlts_test/logs",
  "aip-location": "/home/archivematica/ewt/dlts_test/aips",
  "archivematica-transfer-source": "ADOC transfer source",
  "xfer-location": "/home/archivematica/ewt/dlts_test/xfer",
  "aipstore-location": "/mnt/amatica/AIPsStore",
  "work-location": "/home/archivematica/ewt/dlts_test/aips/aip_queue"
}
```

The project directory can then be used as the working directory for subsequent `ewt` commands.

### `archive`

Closes out the project by removing file assets and compressing the project logs, metadata, and `config.json` into a `.tgz` archive.

## `rstar`

AIP preparation and validation commands.

### Subcommands

* `prep`
* `validate`

## `sip`

SIP generation, scanning, sizing, and validation commands.

### Subcommands

* `gen`
* `scan`

  * `av`
  * `clean`
  * `detox`
* `size`
* `validate`

### `gen`

Generates a SIP from the source material.

### `scan`

Scans the SIP for various conditions. `scan` has its own subcommands for individual types of scans.

#### `scan av`

Scans the SIP for viruses and other malware. Creates a log in the logs directory that the validation step uses

#### `scan clean`

Scans the SIP for common unwanted files, such as `.DS_Store`, `Thumbs.db`, and `Desktop.ini`.

#### `scan detox`

Scans filenames for characters that may cause problems in downstream systems.

### `size`

Prints the size and number of files in the SIP.

```text
$ ewt sip size
ewt sip size, version v1.2.0-alpha
/home/don/ewt-test/dlts_test100/sip: 7 files in 3 directories, 45 MB
```

### `validate`

Validates the SIP.

```$ ewt sip validate
ewt sip validate, v1.2.0-alpha
  * validating SIP at /home/don/ewt-test/dlts_tes100/sip
    1. checking that SIP location exists and is a directory:  OK
    2. checking that SIP directory contains a metadata directory: OK
    3. checking that a valid workorder file exists: OK
    4. checking workorder dlts_test101_aspace_wo.tsv for duplicate cuids: OK
    5. checking all ER directories in workorder exist: OK
    6. checking that there are no extra directories or files in SIP directory: OK
    7. checking that valid transfer-info.txt exists: OK
    8. checking clamscan log for infected files: ERROR, contains 1 infected files
    9. checking detox log: WARNING detox scan contains unresolved filename issues
  * Validation report written to logs/dlts_tes100-sip-validate.log
  * SIP HAS ERRORS
  * SIP HAS WARNINGS
```

## `source`

Commands for working with the source material as configured in the project init step.

### Subcommands

* `size`

print out the size and number of files in source directory
```
$ ewt source size
ewt source size, version v1.2.0-alpha
/home/don/ewt-test-data/dlts_test101/to_rstar/: 7 files in 3 directories, 45 MB
```
* `transfer`

rsync or robocopy the files from from source directory to sip directory

```
$ ewt source transfer, version v1.2.0-alpha
  * Transferring /home/don/ewt-test-data/dlts_test101/to_rstar/ to sip directory
  * Transfer complete
```

## `version`

Prints the installed version of `ewt`.

```text
$ ewt version
ewt v1.2.0
```
