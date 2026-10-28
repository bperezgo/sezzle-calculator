.PHONY: install install-hooks backend-check frontend-check check cover up up-obs down logs smoke

install:
	cd frontend && pnpm install
	cd backend && go mod download

# Installs lefthook if missing, then wires the git hooks (lefthook.yml).
install-hooks:
	@command -v lefthook >/dev/null 2>&1 || go install github.com/evilmartians/lefthook@latest
	lefthook install

backend-check:
	$(MAKE) -C backend check

frontend-check:
	cd frontend && pnpm check

# Everything that must be green before a commit (mirrors CI).
check: backend-check frontend-check

# Coverage reports: backend/coverage.html and frontend/coverage/index.html.
cover:
	$(MAKE) -C backend cover-html
	cd frontend && pnpm test:coverage

up:
	docker compose up --build --detach --wait

# The app plus local logs in Grafana (http://localhost:3001); see README "Local observability".
up-obs:
	docker compose --profile observability up --build --detach --wait

# The profile flag also removes the observability containers when they are running.
down:
	docker compose --profile observability down

# Real calculations against the running stack (`make up` first); same script as CI.
smoke:
	scripts/smoke.sh

logs:
	docker compose logs -f
