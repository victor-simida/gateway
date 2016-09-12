#!/usr/bin/env bash

SERVER=gateway

pid=$(ps -ef | grep "$SERVER$" | grep -v 'grep' | awk '{ print $2 }')
if [ ! -z $pid ]
then
        echo "kill $pid"
        kill $pid
fi

nohup ./$SERVER > ./output_${SERVER}.log 2>&1 &
