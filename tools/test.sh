#!/usr/bin/env bash

Server="http://z1.zhengyinyong.com:9614"

for i in `seq 1 10000`; do
curl --request POST \
  --url "$Server/authorize" \
  --header 'Content-Type: application/json' \
  --data '{"ch_name": "test","en_name": "test","password": "123456"}'

curl --request GET --url "$Server/country"
done