#!/bin/bash
#脚本目录
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd "$SCRIPT_DIR"
[ -x ./gaussdb_exporter ] || chmod +x ./gaussdb_exporter

#pid
mypid=`ps -ef |grep -w '\./gaussdb_exporte[r]'| awk '{print $2}'`
if [ -n "$mypid" ]; then
    echo "gaussdb_exporter running, pid:" $mypid
    exit 0
else
    nohup ./gaussdb_exporter --web.listen-address=:9103 2>&1 &
    echo "starting..."
fi

