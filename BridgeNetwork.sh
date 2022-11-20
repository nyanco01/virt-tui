#!/bin/bash
cd `dirname $0`

SHELLS=`echo ./tmp/shell/*`

for i in $SHELLS
do
    sudo bash $i
done
