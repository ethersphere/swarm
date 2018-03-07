.PHONY:	wolkdb wolk test enssimulation dbchunkstore server keymanager bplus hashdb

wolkdb:	
	@echo "compiling wolkdb server..."
	go build -a -o ./server/wolkdb ./server/wolkdb.go

test:
	@echo "test all."
	@echo "test swarmdb."
	-go test ./swarmdb_test.go
	@echo "test enssimulation."
	-go test ./enssimulation_test.go
	@echo "test dbchunkstore."
	-go test ./dbchunkstore_test.go
	@echo "test keymanager."
	-go test ./keymanager_test.go
	@echo "test bplus."
	-go test ./bplus_test.go
	@echo "test query."
	-go test ./query_test.go
	@echo "test hashdb."
	-go test ./hashdb_test.go
	@echo "test netstats."
	-go test ./netstats_test.go

enssimulation:
	@echo "test enssimulation."
	go test ./enssimulation_test.go

dbchunkstore:
	@echo "test dbchunkstore."
	go test ./dbchunkstore_test.go

keymanager:
	@echo "test keymanager."
	go test ./keymanager_test.go

bplus:
	@echo "test bplus."
	go test bplus_test.go

query:
	@echo "test query."
	go test ./query_test.go

swarmdb:
	@echo "test swarmdb."
	go test ./swarmdb_test.go

hashdb:
	@echo "test hashdb."
	go test hashdb_test.go

netstat:
	@echo "test netstats."
	go test ./netstats_test.go
