





boltdbweb:
	# go get github.com/evnix/boltdbweb
	boltdbweb --db-name=my.db --port=8089

dump:
	bbolt dump my.db 0 # 打印第0个page的数据

bbolt:
	go get go.etcd.io/bbolt/... # 安装bbolt命令，bolt cli客户端

buckets:
	bbolt buckets my.db

# 打印 hex 数据
hexdump:
	hexdump -C my.db
