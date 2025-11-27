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

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

gateway-server:
	go run internal/gateway/main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/techschool/simplebank/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go github.com/techschool/simplebank/worker TaskDistributor

proto:
	@echo "Compiling proto files for all services..."
	@make proto-agent
	@make proto-plan
	@make proto-implementation
	@make proto-diagram
	@make proto-analyzer
	@echo "All proto files compiled successfully!"

proto-agent:
	@echo "Compiling Agent Service proto..."
	rm -f services/agent/pb/*.go
	protoc --proto_path=services/agent/pb \
       --go_out=services/agent/pb --go_opt=paths=source_relative \
       --go-grpc_out=services/agent/pb --go-grpc_opt=paths=source_relative \
       services/agent/pb/*.proto

proto-plan:
	@echo "Compiling Plan Service proto..."
	rm -f services/plan/pb/*.go
	protoc --proto_path=services/plan/pb \
       --go_out=services/plan/pb --go_opt=paths=source_relative \
       --go-grpc_out=services/plan/pb --go-grpc_opt=paths=source_relative \
       services/plan/pb/*.proto

proto-implementation:
	@echo "Compiling Implementation Service proto..."
	rm -f services/implementation/pb/*.go
	protoc --proto_path=services/implementation/pb \
       --go_out=services/implementation/pb --go_opt=paths=source_relative \
       --go-grpc_out=services/implementation/pb --go-grpc_opt=paths=source_relative \
       services/implementation/pb/*.proto

proto-diagram:
	@echo "Compiling Diagram Service proto..."
	rm -f services/diagram/pb/*.go
	protoc --proto_path=services/diagram/pb \
       --go_out=services/diagram/pb --go_opt=paths=source_relative \
       --go-grpc_out=services/diagram/pb --go-grpc_opt=paths=source_relative \
       services/diagram/pb/*.proto

proto-analyzer:
	@echo "Compiling Analyzer Service proto..."
	rm -f services/analyzer/pb/*.go
	protoc --proto_path=services/analyzer/pb \
       --go_out=services/analyzer/pb --go_opt=paths=source_relative \
       --go-grpc_out=services/analyzer/pb --go-grpc_opt=paths=source_relative \
       services/analyzer/pb/*.proto

# 각 서비스 실행
run-gateway:
	go run internal/gateway/main.go

run-agent:
	go run services/agent/main.go

run-plan:
	go run services/plan/main.go

run-implementation:
	go run services/implementation/main.go

run-diagram:
	go run services/diagram/main.go

run-analyzer:
	go run services/analyzer/main.go

# 모든 서비스 동시 실행 (개발용)
run-all:
	@echo "Starting all microservices..."
	@make -j6 run-gateway run-agent run-plan run-implementation run-diagram run-analyzer

evans:
	evans --host localhost --port 9090 -r repl

evans-plan:
	evans --host localhost --port 9091 -r repl

evans-implementation:
	evans --host localhost --port 9092 -r repl

evans-diagram:
	evans --host localhost --port 9093 -r repl

evans-analyzer:
	evans --host localhost --port 9094 -r repl

.PHONY: network postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 new_migration db_docs db_schema sqlc test server mock proto proto-agent proto-plan proto-implementation proto-diagram proto-analyzer evans evans-plan evans-implementation evans-diagram evans-analyzer redis run-gateway run-agent run-plan run-implementation run-diagram run-analyzer run-all