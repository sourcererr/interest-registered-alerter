all:
	go get
	go build

setup:
	docker-compose up -d

run:
	make all
	make setup
	./interest-registered-alerter
