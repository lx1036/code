

# reload prometheus 进程： curl -X POST localhost:9090/-/reload


install:
	# 或者直接 `prometheus --config.file=./prometheus.yml`
	/usr/local/bin/prometheus --config.file=/etc/prometheus/prometheus.yml --storage.tsdb.path=/data/prometheus --storage.tsdb.retention=15d --storage.tsdb.min-block-duration=2h --storage.tsdb.max-block-duration=2h --web.console.libraries=/etc/prometheus/console_libraries --web.console.templates=/etc/prometheus/consoles --web.enable-lifecycle --web.enable-admin-api --web.listen-address=0.0.0.0:9090

