#!/bin/bash

cd /go/obd && ./tracker_server --trackerConfigPath "/go/conf.tracker.ini" &

sleep 1

cd /go/obd  && ./obdserver --configPath "/go/conf.ini"
