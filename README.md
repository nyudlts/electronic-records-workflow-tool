# electronic-records-workflow-tool

## build
<pre>go build -o erwt main.go</pre>

## commands
<pre>
Available Commands:
  aip         erwt aip commands
  amatica     erwt archivematica commands
  aspace      erwt archivesSpace commands
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  project     erwt project commands
  sip         erwt sip commands
  source      erwt source commands
  version     print the version of erwt
</pre>

## Sub Commands

### aip
<pre>
Usage:
  erwt aip [command] [flags]

Available Commands:
  prep        Prepare a list of AIPs for transfer to R*
  size        Get the file count and size of an AIP package
  transfer    Transfer processed AIPS to R*
  validate    Validate AIPS prior to transfer to R*
</pre>
#### aip prep 
<pre>
Prepare a list of AIPs for transfer to R*

Usage:
   ewwt aip prep [flags]

Flags:
      --aip-file string       the location of the aip-file containing aips to process (default finds aipfile in /logs directory)
      --aip-location string   location to stage aips (default "aips/")
  -h, --help                  help for prep
      --tmp-location string   location to store tmp bag-info.txt (default "logs")
</pre>
#### aip size
#### aip transfer
#### aip validate
### amatica
### aspace
### help
print the help message
### project
### sip
### source
### version
print the version of erwt