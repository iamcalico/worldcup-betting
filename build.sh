#!/usr/bin/env bash
go build main.go error_response.go user_authorize.go
mkdir -p release
mv main release/worldcup-betting
cp config.toml release/
cp run.sh release/