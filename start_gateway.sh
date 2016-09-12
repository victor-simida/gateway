#!/bin/bash
BINFILE=gateway

BIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd $BIN_DIR
MONITOR_LOG="$BIN_DIR/monitor.log"
MONITOR_PIDFILE="$BIN_DIR/monitor.pid"
MONITOR_PID=0
if [[ -f $MONITOR_PIDFILE ]]; then
  MONITOR_PID=`cat $MONITOR_PIDFILE`
fi
PIDFILE="$BIN_DIR/$(basename $BINFILE).pid"
PID=0
if [[ -f $PIDFILE ]]; then
  PID=`cat $PIDFILE`
fi

START_CMD=$BIN_DIR/$BINFILE > $BIN_DIR/out_$BINFILE
STOP_CMD="kill $PID"
MONITOR_INTERVAL=5

running() {
  if [[ -z $1 || $1 == 0 ]]; then
    return 1
  fi
  if [[ ! -d /proc/$1 ]]; then
      return 1
  fi
}

start_app() {
  echo "starting $BINFILE"
  nohup $START_CMD &
  if ! $(running $!) ; then
    echo "failed to start $BINFILE"
    exit 1
  fi
  PID=$!
  echo $! > $PIDFILE
  echo "new pid $!"
}

stop_app() {
  if ! $(running $PID) ; then
    return
  fi
  echo "stopping $PID of $BINFILE ..."
  $STOP_CMD
  while $(running $PID) ; do
    sleep 1
  done
}

start_monitor() {
  while [ 1 ]; do
    if ! $(running $PID) ; then
      echo "$(date '+%Y-%m-%d %T') $BINFILE is gone" >> $MONITOR_LOG
      start_app
      echo "$(date '+%Y-%m-%d %T') $BINFILE started" >> $MONITOR_LOG
    fi
    sleep $MONITOR_INTERVAL
  done &
  MONITOR_PID=$!
  echo "monitor pid $!"
  echo $! > $MONITOR_PIDFILE
}

stop_monitor() {
  if ! $(running $MONITOR_PID) ; then
    return
  fi
  echo "stopping $MONITOR_PID of $BINFILE monitor ..."
  kill $MONITOR_PID
  while $(running $MONITOR_PID) ; do
    sleep 1
  done
}

start() {
  start_app
  start_monitor
}

stop() {
  stop_monitor
  stop_app
}

restart() {
  stop
  start
}

restart

