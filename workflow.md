## Workflow Steps
1. Copy from ACMBornDigital to AmaticaStaging using rsync or robocopy
2. Run Clamscan antivirus against package `adoc clamscan`
3. Add transfer-info.txt and work order to metadata directory.
4. Validate the SIP `adoc validate-sip`
5. Stage the SIP for transfer to archivematica `adoc stage`
6. Transfer the SIP to Archivmatica `adoc transfer-am`
7. Prepare AIPs for transfer to R* `adoc prep-list`
8. Validate AIPs for transfer to R* `adoc validate-aip`
9. Transfer AIPS to R* `adoc transfer-rs`
10. Generate check tsv after trasfer to r* completes `adoc check`