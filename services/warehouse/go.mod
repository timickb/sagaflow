module github.com/timickb/sagaflow/services/warehouse

go 1.26.1

replace github.com/timickb/sagaflow/proto => ../../proto

replace github.com/timickb/sagaflow/lib => ../../lib

require (
	github.com/google/uuid v1.6.0
	github.com/segmentio/kafka-go v0.4.50
	github.com/timickb/sagaflow/lib v0.0.0
	github.com/timickb/sagaflow/proto v0.0.0
	google.golang.org/grpc v1.80.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/golang-migrate/migrate/v4 v4.19.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pierrec/lz4/v4 v4.1.16 // indirect
	github.com/rs/zerolog v1.35.0 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120221211-b8f7ae30c516 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
)
