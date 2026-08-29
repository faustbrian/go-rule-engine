.PHONY: competitors docs integration

competitors:
	./verification/check-competitors.sh "$(BENCH_TIME)"

docs:
	./verification/check-docs.sh

integration:
	./verification/check-integrations.sh
