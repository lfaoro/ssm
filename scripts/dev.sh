#!/bin/bash
#
# Copyright (c) 2025 Leonardo Faoro & authors
# SPDX-License-Identifier: MIT

APP=ssm
APP_PATH=./
PID_PATH=/tmp/${APP}_dev.pid
WATCHER_PID=""
CLEANING=false

echo $$ > ${PID_PATH}

cleanup() {
	# prevent recursive cleanup calls
	if $CLEANING; then
		return
	fi
	CLEANING=true

	reset
	echo "cleaning up..."

	trap - SIGINT SIGTERM SIGQUIT

	# kill watcher by PID (not pkill, avoids killing unrelated processes)
	if [[ -n "$WATCHER_PID" ]] && kill -0 "$WATCHER_PID" 2>/dev/null; then
		kill "$WATCHER_PID" 2>/dev/null || true
		wait "$WATCHER_PID" 2>/dev/null || true
	fi

	# kill any running ssm instances started by this script
	pkill -TERM -x "$APP" 2>/dev/null || true
	sleep 0.5
	pkill -9 -x "$APP" 2>/dev/null || true

	rm -f "$PID_PATH"
	exit 0
}
trap cleanup SIGINT SIGTERM SIGQUIT

start_app() {
	reset
	export TERM=xterm-256color
	echo "starting ${APP}"
	go run -ldflags="" ${APP_PATH} --debug
	echo "${APP} process exited"
}

# start background file watcher
start_watcher() {
	if command -v inotifywait &>/dev/null; then
		echo "watching with inotifywait..."
		(
			while true; do
				inotifywait -q -r -e modify,create,delete ${APP_PATH} --include '\.go$'
				echo "file change detected!"
				pkill -TERM -x ${APP} 2>/dev/null || true
			done
		) &
		WATCHER_PID=$!
	elif command -v fswatch &>/dev/null; then
		echo "watching with fswatch..."
		(
			fswatch -r -e '\.go$' ${APP_PATH} | while read -r; do
				echo "file change detected!"
				pkill -TERM -x ${APP} 2>/dev/null || true
			done
		) &
		WATCHER_PID=$!
	else
		echo "watching with polling fallback (install inotifywait or fswatch for better performance)..."
		(
			last_mod=$(find ${APP_PATH} -name '*.go' -exec stat -c '%Y' {} + 2>/dev/null | sort -rn | head -1 || echo 0)
			while true; do
				sleep 1
				cur_mod=$(find ${APP_PATH} -name '*.go' -exec stat -c '%Y' {} + 2>/dev/null | sort -rn | head -1 || echo 0)
				if [[ "$cur_mod" != "$last_mod" ]]; then
					echo "file change detected!"
					last_mod=$cur_mod
					pkill -TERM -x ${APP} 2>/dev/null || true
				fi
			done
		) &
		WATCHER_PID=$!
	fi
}

start_watcher
start_app

# main loop - after app exits, restart it
while true; do
	echo "restarting ${APP}"
	sleep 0.5
	start_app
done
