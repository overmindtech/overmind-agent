#!/bin/bash
############################################################################
#
# Usage: loc7.sh [options] file ...
#
# Count the number of lines in a given list of files.
# Uses a for loop over all arguments.
#
# Options:
#  -h     ... help message
#  -d n ... consider only files modified within the last n days
#  -w n ... consider only files modified within the last n weeks
#
# Limitations: 
#  . only one option should be given; a second one overrides
#
############################################################################

l=0
n=0
s=0
for f in "/etc/passwd" "/etc/hosts"
do
  # do the line count
  l=`wc -l < $f`
  echo "$f: $l"
  # increase the counters
  n=$[ $n + 1 ]
  s=$[ $s + $l ]
done

echo "$n files in total, with $s lines in total"