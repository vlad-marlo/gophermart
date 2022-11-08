.PHONY: build
build:
	go build -v ./cmd/gophermart
	cd cmd/gophermart && go build -o gophermart

.PHONY: test
test:
	go clean -testcache
	go test -cover -v ./internal/store/sqlstore
	go test -cover -v ./pkg/...
	go test -cover -v ./internal/server


.PHONY: t
t:
	./gophermarttest \
                  -test.v -test.run=^TestGophermart$ \
                  -gophermart-binary-path=cmd/gophermart/gophermart \
                  -gophermart-host=localhost \
                  -gophermart-port=8000 \
                  -gophermart-database-uri="postgresql://marlo:819819@localhost:5432/gophermart_test?sslmode=disable" \
                  -accrual-binary-path=cmd/accrual/accrual_linux_amd64 \
                  -accrual-host=localhost \
                  -accrual-port=5000 \
                  -accrual-database-uri="postgresql://marlo:819819@localhost:5432/gophermart_test?sslmode=disable"


.DEFAULT_GOAL := build
