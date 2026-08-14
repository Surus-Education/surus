# Thin alias over the pnpm scripts in package.json, kept only so `make up` still
# works for anyone with the muscle memory. `pnpm dev` is the real entry point and
# is the one documented in the README.
#
# The previous version ran `nohup sh -c 'cd api && go run ./cmd/server'`, which
# is broken on Windows: make hands the command to a Unix shell whose PATH has no
# Windows `go` or `node`, so both services failed with "go: not found" /
# "exec: node: not found". concurrently spawns the processes directly instead.
#
# It also ran `docker compose up -d`. The project runs on Neon now, so that
# started a Postgres nothing connects to. docker-compose.yaml is still in the
# repo as a fallback — start it yourself if you want a local database.

.PHONY: up dev down

up dev:
	pnpm dev

down:
	@echo "Ctrl-C the 'pnpm dev' process — concurrently (-k) stops both services."
	@echo "For the optional local Postgres: docker compose down"
