# adoc-preprocess
Prep Tool for Electronic Records being transferred to R* through Archivematica

## commands
<pre>
Available Commands:
  check         Check that DOs exist in ArchivesSpace
  clamscan      Run clamav against a package
  completion    Generate the autocompletion script for the specified shell
  help          Help about any command
  prep-list     Prepare a list of AIPs for transfer to R*
  prep-single   Prepare a single AIP for transfer to R*
  stage         pre-process SIPs
  transfer-am   Transfer SIPS to R*
  transfer-rs   Transfer processed AIPS to R*
  validate-aips Validate AIPS prior to transfer to R*
  validate-sip  validate sips prior to transfer to Archivematica
  version       print the version of adoc
</pre>
### check
Check that DOs exist in Archivesspace
<pre>
Usage:
   check [flags]
Flags:
      --aspace-config string
      --aspace-environment string
      --aspace-workorder string
  -h, --help                        help for check
</pre>
### clamscan
Run clamav against a package
<pre>
Usage:
   clamscan [flags]
Flags:
  -h, --help                      help for clamscan
      --regexp string
      --staging-location string
</pre>
### help
print the help message
### prep-list
Prepare a list of AIPs for transfer to R*
<pre>
Usage:
   prep-list [flags]
Flags:
      --aip-file string
  -h, --help                      help for prep-list
      --staging-location string
      --tmp-location string
</pre>
### prep-single
### stage 
### transfer-am
Transfer SIPS to R*
<pre>
Usage:
   transfer-am [flags]
Flags:
      --config string           if not set will default to `/home/'username'/.config/go-archivematica.yml
  -h, --help                    help for transfer-am
      --poll int                pause time, in seconds, between calls to Archivematica api to check status (default 5)
      --regexp string           regular expression to filter directory names to transfer to Archivmatica (default ".*")
      --xfer-directory string   Location of directories top transfer to Archivematica (required)
</pre>
### transfer-rs
### validate-aips
### validate-sips
### version