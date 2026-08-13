# electronic-records-workflow-tool v1.2.0

## Prerequisites
* `go v1.26`
* `clamdscan`
* `detox`
* `rsync`

## Configs
configurations for ewt are stored in a .config directory in the users home directory. 

* ~/.config/ewt.config -> general ewt config
* ~/.config/go-archivematica.yml -> configuratin for go-archivematica library
* ~/.conifg/go-aspace.yml -> configuration for go-aspace-library

samples of these configs can be found in th `templates` directory
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

### `amatica`

Archivematica-related commands.

#### Subcommands

* `prep`

move and update all directories in `sip` to `xfer`
```
$ ewt amatica prep
ewt amatica prep, v1.2.0
  * WORKER 1 processing DLTS_TEST_101_ER_1
    source: /home/don/ewt-test/dlts_tes100/sip/DLTS_TEST_101_ER_1
    target: xfer/dlts_tes100_DLTS_TEST_101_ER_1/DLTS_TEST_101_ER_1
  * WORKER 1 completed DLTS_TEST_101_ER_1
```
* `transfer`
* `gen`

copy a processingMCP.xml from the templates directory to each package in the `xfer` directory.


### `aspace`

ArchivesSpace-related commands.

#### Subcommands

* `check`

check all dos in workorder exist in archivesspace and write the ouput to a tsv file in the logs directory

```
$ ewt aspace check
ewt aspace check, v1.2.0
Checking /repositories/6/archival_objects/1059314: OK
Checking /repositories/6/archival_objects/1059315: OK
Checking /repositories/6/archival_objects/1059316: OK
Checking /repositories/6/archival_objects/1059317: OK
aspace checkfile written to: logs/nyuarchives_rg35_5-aspace-check.tsv
```

### `completion`

Generates the autocompletion script for the specified shell.

### `help`

Displays help information for `ewt` and its commands.

### `project`

Project initialization and closeout commands.

#### `init`

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

#### `archive`

Closes out the project by removing file assets and compressing the project logs, metadata, and `config.json` into a `.tgz` archive.

### `rstar`

AIP preparation and validation commands.

#### Subcommands

* `prep`
* `validate`

### `sip`

SIP generation, scanning, sizing, and validation commands.

#### Subcommands

* `gen`
  * `transfer`
* `scan`
  * `av`
  * `clean`
  * `detox`
* `size`
* `validate`

#### `gen`

generate metadata files

##### `gen transfer`
generate a transfer-info.txt file at ./sip/metadata
```$ ewt sip gen transfer -p dm
ewt sip gen transfer, version v1.2.0-alpha
  * generating transfer info for profile: dm
```  

#### `scan`

Scans the SIP for various conditions. `scan` has its own subcommands for individual types of scans.

##### `scan av`

Scans the SIP for viruses and other malware. Creates a log in the logs directory that the validation step uses.

```$ ewt sip scan av
ewt sip scan av,  v1.2.0-alpha
 *  scanning sip/metadata/transfer-info.txt
 *  scanning sip/DLTS_TEST_101_ER_1/eicar.com
 *  scanning sip/DLTS_TEST_101_ER_1/fales_mss.657.json
 *  scanning sip/DLTS_TEST_101_ER_1/fales_mss.657 - Copy.json
 *  scanning sip/metadata/dlts_test101_aspace_wo.tsv
 *  1 error(s) encountered:
    [ERROR] clamdscan malware detected: /home/don/ewt-test/dlts_tes100/sip/DLTS_TEST_101_ER_1/eicar.com
```

##### `scan clean`

Scans the SIP for common unwanted files, such as `.DS_Store`, `Thumbs.db`, and `Desktop.ini` and deletes them. Creates a log in the logs directory. 

```$ ewt sip scan clean
ewt sip clean, version v1.2.0-alpha
  * deleted "/home/don/ewt-test/dlts_tes100/sip/DLTS_TEST_101_ER_1/.DS_Store"
  * deleted "/home/don/ewt-test/dlts_tes100/sip/DLTS_TEST_101_ER_1/Icon\r"
  * 2 files deleted
```

##### `scan detox`

Scans filenames for characters that may cause problems in downstream systems.

```
$ ewt sip scan detox
ewt sip scan detox, v1.2.0-alpha
  * detox chars found:
    /home/don/ewt-test/dlts_tes100/sip/DLTS_TEST_101_ER_1/fales_mss.657 - Copy.json -> /home/don/ewt-test/dlts_tes100/sip/DLTS_TEST_101_ER_1/fales_mss.657-Copy.json
```

#### `size`
Prints the size and number of files in the SIP.

```text
$ ewt sip size
ewt sip size, version v1.2.0-alpha
/home/don/ewt-test/dlts_test100/sip: 7 files in 3 directories, 45 MB
```

#### `validate`

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

### `source`

Commands for working with the source material as configured in the project init step.

#### Subcommands

##### `source size`

print out the size and number of files in source directory
```
$ ewt source size
ewt source size, version v1.2.0-alpha
/home/don/ewt-test-data/dlts_test101/to_rstar/: 7 files in 3 directories, 45 MB
```
##### `source transfer`

rsync or robocopy the files from from source directory to sip directory

```
$ ewt source transfer, version v1.2.0-alpha
  * Transferring /home/don/ewt-test-data/dlts_test101/to_rstar/ to sip directory
  * Transfer complete
```

### `version`

Prints the installed version of `ewt`.

```text
$ ewt version
ewt v1.2.0
```
