#!/bin/sh


while true;
do
	go build -gcflags="all=-N -l" -o main
	debug_cmd="dlv --listen=:2345 --headless=true --api-version=2 exec ./main"
	while  [ $# -gt 0 ]
	do
			echo $1

			parameters="$parameters -- $1"
			#　左移一个参数，这样可以使用$1遍历所有参数
			shift
	done

	cmd="$debug_cmd $parameters"
	eval $cmd

done

