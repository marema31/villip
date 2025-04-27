.PHONY: all start run stop test k8sup k8sdown

all: test

default: help

help: ## show help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:  make [target]\033[36m\033[0m\n\n"} /^[$$()% 0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

start: ## start end 2 end environment (mock and applications)
	./start.sh

run: ## run the test suite
	docker-compose start tcpclient
	docker-compose start venom

stop: ## stop end 2 end environment (mock and applications)
	docker-compose stop

test: start run stop ## Start environment, run test and tear down apllication and mocks
	[ `\grep -c  failure $(CURDIR)/outputs/test_results.xml` -eq 0 ]

k8sup:  ## spin up end 2 end environment (kind cluster)
	./k8s.sh

k8sdown: ## tear down end 2 end environment (kind cluster)
	kind delete cluster --name villip
