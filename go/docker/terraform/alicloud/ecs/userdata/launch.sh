#!/usr/bin/env bash
apt-get update -y && apt-get install -y nginx > /tmp/nginx.log
cd / && mkdir -p alicloud/userdata
