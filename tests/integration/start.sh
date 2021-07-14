#!/bin/bash

# TODO: Replace with healthcheck
sleep 4

cd /obd && ./tracker_server --trackerConfigPath "/obd/conf.tracker.ini" &

sleep 1

/obd/obdserver --configPath "/obd/conf.ini.alice" &
/obd/obdserver --configPath "/obd/conf.ini.bob"