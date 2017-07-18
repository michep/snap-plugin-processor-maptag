asmdisks=$(oracleasm listdisks)

for disk in $asmdisks; do
  outstr=$(oracleasm querydisk -p $disk |grep '^\/')
  if [[ -z $outstr ]]; then
    continue
  fi
  dev=$(echo $outstr |grep -oP '^.+?(?=:)')
  label=$(echo $outstr |grep -oP '(?<=LABEL=").+?(?=")')
  printf "%s\t%s\n" $dev $label
done
