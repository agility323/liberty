#!/bin/sh
nohup ./gate --test_ip=192.168.229.205 --test_port=7890 --test_etcd="{\"1\":1}" --conf=config_c1.json > log1 &
