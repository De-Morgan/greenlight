

#psql --host=localhost --dbname=greenlight-db --username=greenlight-admin
greenlightdsn:
	export GREENLIGHT_DB_DSN="postgres://greenlight:password@localhost/greenlight-db?sslmode=disable"
migrateup:
	migrate -database ${GREENLIGHT_DB_DSN} -path db/migrations up