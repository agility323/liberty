#!/bin/bash
nohup ./login --conf=config.json > log &
tail -f log
