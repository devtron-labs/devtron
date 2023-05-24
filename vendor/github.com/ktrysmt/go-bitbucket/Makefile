.DEFAULT_GOAL := help

env: ## check env for e2e testing
ifndef BITBUCKET_TEST_USERNAME
	$(error `BITBUCKET_TEST_USERNAME` is not set)
endif
ifndef BITBUCKET_TEST_PASSWORD
	$(error `BITBUCKET_TEST_PASSWORD` is not set)
endif
ifndef BITBUCKET_TEST_OWNER
	$(error `BITBUCKET_TEST_OWNER` is not set)
endif
ifndef BITBUCKET_TEST_REPOSLUG
	$(error `BITBUCKET_TEST_REPOSLUG` is not set)
endif

test: env ## run go test all
	go test -v ./tests

test/swagger:
	env BITBUCKET_API_BASE_URL=http://0.0.0.0:4010 go test -v ./...

help: ## print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test test/swagger help
