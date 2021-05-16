#!/bin/bash

# TODO: Replace with healthcheck
sleep 4

cd /obd && ./tracker_server --trackerConfigPath "/obd/conf.tracker.ini" &

sleep 1

cd /obd  && ./obdserver --configPath "/obd/conf.ini"