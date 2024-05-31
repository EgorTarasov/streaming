.PHONY: build-framer
build-framer:
	docker build ./framer/ -t framer

.PHONY: run-framer
run-framer: build-framer
	docker run -d framer


.PHONY: build-detection
build-detection:
	docker build ./detection/ -t detection

.PHONY: run-detection
run-detection: build-detection
	docker run -d -e PREDICTION_TOPIC="prediction" -e IMAGES_TOPIC="frames-splitted"  -e KAFKA_BROKERS="localhost:9091" --name test detection

.PHONY: set-up-kafka
set-up-kafka:
	docker exec -it zookeeper ./bin/kafka-topics.sh --create --topic <topic_name> --bootstrap-server <broker_address> [--partitions <number_of_partitions>] [--replication-factor <replication_factor>] [--config <key=value>]


.PHONY: format
format:
	$(info Running goimports...)
	test -f ${SMART_IMPORTS} || GOBIN=${LOCAL_BIN} go install github.com/pav5000/smartimports/cmd/smartimports@latest
	${SMART_IMPORTS} -exclude pkg/,internal/pb  -local 'gitlab.ozon.dev'

ifeq (,$(wildcard .env))
    # Если файл .env отсутствует, устанавливаем параметры по умолчанию
    POSTGRES_USER := 'etarasov'
    POSTGRES_PASSWORD := 'Do-Megor2023'
    POSTGRES_DB := 'dev'
    POSTGRES_HOST := '192.168.1.70'
    POSTGRES_PORT := 54002
else
    # Иначе, подключаем переменные из файла .env
    include .env
    export
endif
POSTGRES_SETUP_TEST := user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} host=${POSTGRES_HOST} port=${POSTGRES_PORT} sslmode=disable

INTERNAL_PKG_PATH=$(CURDIR)/internal
MIGRATION_FOLDER=$(CURDIR)/migrations

.PHONY: migration-create
migration-create:
	goose -dir "$(MIGRATION_FOLDER)" create "$(name)" sql

.PHONY: migration-up
migration-up:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" up


.PHONY: migration-down
migration-down:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" down

