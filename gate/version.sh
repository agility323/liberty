#!/bin/sh
date -d "$(git log --pretty=format:'%ci' | head -1)" +"%Y%m%d_%H%M%S"
