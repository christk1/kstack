#!/usr/bin/env bash
set -eu

# Simple helper to start/stop background kubectl port-forwards for the demo services
# Logs are written to /tmp/kstack-portforward-logs and PIDs to .pid files there.

ACTION=${1:-}
LOGDIR=/tmp/kstack-portforward-logs

case "$ACTION" in
	stop)
		if [ -d "$LOGDIR" ]; then
			echo "Stopping port-forwards (reading PIDs from $LOGDIR)"
			for pidfile in "$LOGDIR"/*.pid; do
				[ -e "$pidfile" ] || continue
				pid=$(cat "$pidfile" || true)
				if [ -n "${pid:-}" ] && kill -0 "$pid" 2>/dev/null; then
					echo "Killing PID $pid ($pidfile)"
					kill "$pid" 2>/dev/null || true
				fi
				rm -f "$pidfile" || true
			done
			echo "Port-forwards stopped"
		else
			echo "No port-forward log dir found ($LOGDIR)"
		fi
		exit 0
		;;
		start)
		;;
		status)
			if [ -d "$LOGDIR" ]; then
				echo "Current port-forward PIDs:"
				ls -1 "$LOGDIR"/*.pid 2>/dev/null || echo "(none)"
			else
				echo "No port-forward log dir found ($LOGDIR)"
			fi
			exit 0
			;;
	*)
			echo "Usage: $0 [start|stop|status]"
		exit 1
		;;
esac

mkdir -p "$LOGDIR"

echo "Starting port-forwards (logs -> $LOGDIR)"

kubectl port-forward svc/grafana 3000:80 -n monitoring >"$LOGDIR/grafana.log" 2>&1 &
echo $! > "$LOGDIR/grafana.pid"

kubectl port-forward svc/prometheus-server 9090:80 -n monitoring >"$LOGDIR/prometheus.log" 2>&1 &
echo $! > "$LOGDIR/prometheus.pid"

kubectl port-forward svc/example-app-example-app 8080:8080 -n app >"$LOGDIR/example.log" 2>&1 &
echo $! > "$LOGDIR/example.pid"

echo "Port-forwards started"
echo "Grafana -> http://localhost:3000"
echo "Prometheus -> http://localhost:9090"
echo "Example app -> http://localhost:8080"

echo "PIDs:"
ls -1 "$LOGDIR"/*.pid || true

echo "Tail logs with: tail -f $LOGDIR/grafana.log $LOGDIR/prometheus.log $LOGDIR/example.log"
