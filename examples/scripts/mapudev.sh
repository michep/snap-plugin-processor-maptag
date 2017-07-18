#!/bin/sh

asmudev="/dev/oracleasmudev"

if [[ ! -d  $asmudev/disks/ ]]; then
  exit 1
fi

asmdisks=$(ls -d -1 $asmudev/disks/*)
for disk in $asmdisks; do
  dev=$(readlink -f $disk)
  label=$(basename $disk)
  printf "%s\t%s\n" $dev $label
done
