#!/bin/bash
mypid=`ps -ef |grep -w '\./gaussdb_exporte[r]'| awk '{print $2}'`
if [ -z "${mypid}" ];then
    echo "no exporter running..."
    exit 0
else
    echo "gaussdb_exporter pid: "${mypid}", killing"
    kill ${mypid}
    sleep 2
    mypid=`ps -ef |grep -w '\./gaussdb_exporte[r]'| awk '{print $2}'`
    if [ -n "$mypid" ];then
        kill -9 ${mypid}
        sleep 1
    fi
    echo "killed"
fi
