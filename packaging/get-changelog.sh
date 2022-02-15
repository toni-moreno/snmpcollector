#!/bin/bash
set -vx
awk  "/^# $1/{flag=1; next } /^# v/{flag=0} flag" CHANGELOG.md 
