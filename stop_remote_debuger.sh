#!/bin/bash

pids=`ps -ef | grep dlv | grep -v grep | awk '{print $2,$3}'`
echo $pids

arr=($pids)

echo ${arr[0]}
echo ${arr[1]}

if [ ${arr[1]} -ne 1 ]
then
	kill -9 ${arr[1]}
fi

if [ -n ${arr[0]} ]
then
	kill -9 ${arr[0]}
fi


echo 'Stoped.'

