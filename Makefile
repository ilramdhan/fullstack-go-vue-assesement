.DEFAULT_GOAL := help

.PHONY: help up down rebuild logs ps clean be-test fe-test test

help:
	@echo "Stack-level targets:"
	@echo "  make up        - docker compose up --build (BE + FE)"
	@echo "  make down      - docker compose down"
	@echo "  make rebuild   - down, rebuild images, up"
	@echo "  make logs      - tail compose logs"
	@echo "  make ps        - list running services"
	@echo "  make clean     - down + remove volumes (drops the sqlite db)"
	@echo ""
	@echo "Per-package targets (delegated):"
	@echo "  make be-test   - run backend test suite"
	@echo "  make fe-test   - run frontend test suite"
	@echo "  make test      - run both test suites"

up:
	docker compose up --build -d
	@echo
	@echo "Frontend: http://localhost:8088"
	@echo "Backend:  http://localhost:8080/healthz"

down:
	docker compose down

rebuild:
	docker compose down
	docker compose build --no-cache
	docker compose up -d

logs:
	docker compose logs -f --tail=100

ps:
	docker compose ps

clean:
	docker compose down -v

be-test:
	$(MAKE) -C backend test

fe-test:
	cd frontend && npm test -- --run

test: be-test fe-test
