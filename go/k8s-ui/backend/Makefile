
sqlall:
	go run database/syncdb.go orm sqlall > _dev/k8s.sql

initdata:
	go run database/generatedata/main.go > _dev/k8s-data.sql
