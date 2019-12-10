#!/usr/bin/env bash

while true
do
    ps -p 1 -o %mem >> memlog
    sleep 1
done