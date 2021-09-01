.PHONY: all

all: test

start:
	./start.sh

run:
	docker-compose start venom

stop:
	docker-compose stop

test: start run stop
	[ `\grep -c  failure $(CURDIR)/outputs/test_results.xml` -eq 0 ]
