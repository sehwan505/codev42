DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable


image:
	docker build -f deployments/Dockerfile.agent -t codev42-agent-server .
	docker build -f deployments/Dockerfile.gateway -t codev42/gin-gateway .

image-agent:
	docker build -f deployments/Dockerfile.agent -t codev42-agent-server .

image-gateway:
	docker build -f deployments/Dockerfile.gateway -t codev42/gin-gateway .

db-apply:
	atlas migrate apply --env "gorm"

db-diff:
	atlas migrate diff --env "gorm"

db_docs:
	dbdocs build doc/db.dbml

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/techschool/simplebank/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go github.com/techschool/simplebank/worker TaskDistributor

proto:
	rm -f pb/*.go
	protoc --proto_path=pb \
       --go_out=pb --go_opt=paths=source_relative \
       --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
       pb/*.proto

evans:
	evans --host localhost --port 9090 -r repl


.PHONY: network postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 new_migration db_docs db_schema sqlc test server mock proto evans redis