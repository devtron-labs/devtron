PACKAGES=$$(go list ./... | grep -v '/vendor/')

.PHONEY: test
test:
	go test -cover $(PACKAGES)

.PHONEY: bench
bench:
	go test -cover -bench . -benchmem -run 'Benchmark.*' $(PACKAGES)

.PHONEY: vet
vet:
	go vet $(PACKAGES)

deps.svg: *.go
	(   echo "digraph G {"; \
	go list -f '{{range .Imports}}{{printf "\t%q -> %q;\n" $$.ImportPath .}}{{end}}' \
		$$(go list -f '{{join .Deps " "}}' .) .; \
	echo "}"; \
	) | dot -Tsvg -o $@
