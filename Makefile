.PHONY: up down

up:
	docker compose up -d
	nohup sh -c 'cd api && go run ./cmd/server' > .api.log 2>&1 & echo $$! > .api.pid
	nohup sh -c 'cd web && pnpm dev' > .web.log 2>&1 & echo $$! > .web.pid
	@echo "Started. Logs: .api.log  .web.log"

down:
	-kill $$(cat .api.pid 2>/dev/null) 2>/dev/null; rm -f .api.pid
	-kill $$(cat .web.pid 2>/dev/null) 2>/dev/null; rm -f .web.pid
	docker compose down
