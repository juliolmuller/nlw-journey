package main

//go:generate goapi-gen -p spec -o internal/api/spec/journey.gen.spec.go internal/api/spec/journey.spec.json
//go:generate tern migrate -m ./internal/pgstore/migrations/ -c ./internal/pgstore/migrations/tern.conf
//go:generate sqlc generate -f internal/pgstore/sqlc.yml
