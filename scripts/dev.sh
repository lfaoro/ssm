#!/bin/bash

APP=ssm
APP_PATH=./
PID_PATH=/tmp/${APP}_dev.pid
echo $$ > ${PID_PATH}

cleanup() {
    reset
    echo "cleaning up..."

    # prevent recursive cleanup calls
    trap - SIGINT SIGTERM SIGQUIT

    pkill -TERM -x $NOTIFY_PID || pkill -9 -x $NOTIFY_PID || true
    pkill -TERM -x inotifywait || pkill -9 -x inotifywait || true
    pkill -TERM -x "$APP" || pkill -9 -x "$APP" || true

    sleep 1s
    make stop
    exit 0
}
trap cleanup SIGINT SIGTERM SIGQUIT

start_app(){
    reset
    export TERM=xterm-256color
    echo "starting ${APP}"
    go run -ldflags="" ${APP_PATH} --debug
    echo "${APP} process exited"
}

# background file monitoring
(
    while true; do
        inotifywait -q -r -e modify ${APP_PATH} --include '\.go$'
        echo "file change detected!"
        pkill -TERM -x ${APP} 2>/dev/null || true
    done
) &
NOTIFY_PID=$!

# start the app initially
start_app

# main loop - after app exits, restart it
while true; do
    echo "restarting ${APP}"
    sleep 0.5
    start_app
done
