#!/bin/bash

cur_dir=`pwd`

script_name=$1;

if [ -z $script_name ]; then
    echo "请输入要执行的程序名字"
    exit 0
fi

echo $script_name

export GOPATH="$cur_dir/:$GOPATH";

echo $GOPATH

`go run $script_name`
