DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable


image:
	docker build -f deployments/Dockerfile.agent -t codev42/agent-server .
	docker build -f deployments/Dockerfile.gateway -t codev42/gin-gateway .

image-agent:
	docker build -f deployments/Dockerfile.agent -t codev42/agent-server .

image-gateway:
	docker build -f deployments/Dockerfile.gateway -t codev42/gin-gateway .


db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run cmd/codev42/main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/techschool/simplebank/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go github.com/techschool/simplebank/worker TaskDistributor

proto:
	rm -f services/agent/pb/*.go
	protoc --proto_path=services/agent/pb \
       --go_out=services/agent/pb --go_opt=paths=source_relative \
       --go-grpc_out=services/agent/pb --go-grpc_opt=paths=source_relative \
       services/agent/pb/*.proto

evans:
	evans --host localhost --port 9090 -r repl


.PHONY: network postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 new_migration db_docs db_schema sqlc test server mock proto evans redis