#!/bin/sh

cat /etc/fstab |grep ext[2,3,4]| awk '{print($2);}'|xargs -L 1 findmnt |grep '^\/'|awk '{ printf "%s\t%s\n", $2, $1 }'
