# electronic-records-workflow-tool v1.2.0

## build
<pre>go build -o ewt main.go</pre>

## commands
<pre>
Available Commands:
  amatica     ewt Archivematica commands
  aspace      ewt ArchivesSpace commands
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  project     ewt project commands
  rstar       ewt aip commands
  sip         ewt sip commands
  source      ewt source commands
  version     print the version of ewt
</pre>

## Subcommands

### amatica
#### Subcommands
- prep
- transfer
### aspace
#### subcommands
- check
### completion
### help
### project
####
-init flags -c collection_code -s /path/to/package/to/transfer
<pre>
project init creates a new project directory in the current location. The project directory contains all directories to run ewt plus a config.json file which contains all neccessary fileds to run the appplication

$ewt project -c init dlts_test -s /mnt/amatica/testing/ewt-test-summer-2026/to_rstar/
$ cd dlts_test
$ cat config.json
example:
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

</pre>
-archive
remove all file assets and compress logs, metadata and config.json to a .tgz file. 

### rstar
#### Subcommands
- prep
- validate
### sip
#### Subcommands
- gen
- scan 
- size
- validate
### source
- size
- transfer
### version
<pre>
print the version of ewt
$ ewt version
ewt v1.2.0 
</pre> 