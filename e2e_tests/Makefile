.PHONY: all start run stop test k8sup k8sdown

all: test

start:
	./start.sh

run:
	docker-compose start tcpclient
	docker-compose start venom

stop:
	docker-compose stop

test: start run stop
	[ `\grep -c  failure $(CURDIR)/outputs/test_results.xml` -eq 0 ]

k8sup:
	./k8s.sh

k8sdown:
	kind delete cluster --name villip
