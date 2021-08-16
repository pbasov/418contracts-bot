#!/bin/bash
mkdir storage || true

wget https://eve-static-data-export.s3-eu-west-1.amazonaws.com/tranquility/sde.zip -O /tmp/sde.zip
unzip -oj /tmp/sde.zip sde/fsd/typeIDs.yaml -d storage
