#!/bin/sh
BASE_PATH=mautrix_signal/db/upgrade

LATEST_UPGRADE=`grep -hE 'upgrade_v[0-9]' $BASE_PATH/*.py | grep -o '[0-9]*' | sort -n | tail -n1`
INITIAL_UPGRADES_TO=`grep upgrades_to $BASE_PATH/v00_latest_revision.py | grep -o '[0-9]*'`

if [ $LATEST_UPGRADE -ne "$INITIAL_UPGRADES_TO" ]; then
    echo "upgrades_to in $BASE_PATH/v00_latest_revision.py ($INITIAL_UPGRADES_TO) does not match the latest version in $BASE_PATH ($LATEST_UPGRADE)"
    echo "This looks like an error during an upstream take-in. Make sure the numbers match and all conflicts are properly resolved"
    exit 1
fi
