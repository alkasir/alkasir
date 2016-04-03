# Sample data readme


# BGPDump sample data

The sample bgp logs log files are created from randomly selecting output lines from bgpdump:

```
bgpdump -m  .../bview.20150410.gz -m | awk 'BEGIN {srand()} !/^$/ { if (rand() <= .0001) print $0}' > bview.20150410.log

```


the invalid dump sample is edited to cause some parsing errors
