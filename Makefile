DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable


image:
	docker build -f deployments/Dockerfile.plan -t plan-service:latest .
	docker build -f deployments/Dockerfile.implementation -t implementation-service:latest .
	docker build -f deployments/Dockerfile.diagram -t diagram-service:latest .
	docker build -f deployments/Dockerfile.analyzer -t analyzer-service:latest .
	docker build -f deployments/Dockerfile.gateway -t gin-gateway:latest .

image-plan:
	docker build -f deployments/Dockerfile.plan -t plan-service:latest .

image-implementation:
	docker build -f deployments/Dockerfile.implementation -t implementation-service:latest .

image-diagram:
	docker build -f deployments/Dockerfile.diagram -t diagram-service:latest .

image-analyzer:
	docker build -f deployments/Dockerfile.analyzer -t analyzer-service:latest .

image-gateway:
	docker build -f deployments/Dockerfile.gateway -t gin-gateway:latest .

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

run-plan:
	cd services/plan && go run main.go

run-implementation:
	cd services/implementation && go run main.go

run-diagram:
	cd services/diagram && go run main.go

run-analyzer:
	cd services/analyzer && go run main.go

# 모든 서비스 동시 실행 (개발용)
run-all:
	@echo "Starting all microservices..."
	@make -j6 run-gateway run-plan run-implementation run-diagram run-analyzer

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

# Kubernetes / Helm
k8s-install:
	@echo "Installing Codev42  to Kubernetes..."
	cd deployments/codev42 && helm dependency update
	helm install codev42 deployments/codev42 -n codev42 --create-namespace

k8s-upgrade:
	@echo "Upgrading Codev42 ..."
	helm upgrade codev42 deployments/codev42 -n codev42

k8s-uninstall:
	@echo "Uninstalling Codev42 ..."
	helm uninstall codev42 -n codev42

k8s-status:
	@echo "Checking Codev42  status..."
	kubectl get all -n codev42

k8s-logs-plan:
	kubectl logs -f deployment/plan-service -n codev42

k8s-logs-impl:
	kubectl logs -f deployment/implementation-service -n codev42

k8s-logs-diagram:
	kubectl logs -f deployment/diagram-service -n codev42

k8s-logs-analyzer:
	kubectl logs -f deployment/analyzer-service -n codev42

k8s-logs-gateway:
	kubectl logs -f deployment/gin-gateway -n codev42

.PHONY: network postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 new_migration db_docs db_schema sqlc test server mock proto proto-agent proto-plan proto-implementation proto-diagram proto-analyzer evans evans-plan evans-implementation evans-diagram evans-analyzer redis run-gateway run-agent run-plan run-implementation run-diagram run-analyzer run-all k8s-install k8s-upgrade k8s-uninstall k8s-status k8s-logs-plan k8s-logs-impl k8s-logs-diagram k8s-logs-analyzer k8s-logs-gateway