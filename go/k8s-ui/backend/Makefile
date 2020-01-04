
sqlall:
	go run database/syncdb.go orm sqlall > _dev/k8s.sql

# create tables
syncdb:
	go run database/syncdb.go orm syncdb

initdata:
	go run database/generatedata/main.go > _dev/k8s-data.sql
