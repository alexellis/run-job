#!/bin/sh

for f in bin/run-job*; do shasum -a 256 $f > $f.sha256; done
