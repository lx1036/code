
# dump tables ddl into k8s.sql file
sqlall:
	go run database/syncdb.go orm sqlall > _dev/k8s.sql

# create tables in mysql instance
create_tables:
	go run database/syncdb.go orm syncdb

initdata:
	go run database/generatedata/main.go > _dev/k8s-data.sql

db_seed:
	cd _dev && go run db_seed.go
