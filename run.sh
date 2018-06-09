#!/usr/bin/env bash
ps aux | grep worldcup | grep -v grep | awk '{print $2}' | xargs sudo kill -9
./worldcup-betting >> log 2>&1 &