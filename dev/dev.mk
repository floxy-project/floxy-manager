_COMPOSE_BASE=docker compose -f dev/docker-compose.yml --project-name ${NAMESPACE} --env-file dev/platform.env

dev-up: ## Up the environment in docker compose (base services only)
	${_COMPOSE_BASE} up -d

dev-down: ## Down the environment in docker compose
	${_COMPOSE_BASE} down --remove-orphans

dev-clean: ## Down the environment in docker compose with image cleanup
	${_COMPOSE_BASE} down --remove-orphans -v --rmi all

dev-build-proxy: ## Building floxym-reverse-proxy
	${_COMPOSE_BASE} build floxym-reverse-proxy

dev-build-backend: ## Building floxym with SSH key for private repo access
	@if [ ! -f ${HOME}/.ssh/id_rsa ]; then \
		echo "Error: SSH key not found at ${HOME}/.ssh/id_rsa"; \
		echo "Please create an SSH key or use 'make dev-build-backend-token' instead"; \
		exit 1; \
	fi
	DOCKER_BUILDKIT=1 docker build \
		--secret id=ssh_key,src=${HOME}/.ssh/id_rsa \
		--target prod \
		-f Dockerfile \
		-t ${NAMESPACE}_floxym:latest \
		.

dev-build-backend-token: ## Building floxym with GitHub token (set GITHUB_TOKEN in compose.env)
	@if [ -z "$$GITHUB_TOKEN" ] && [ -f dev/compose.env ]; then \
		GITHUB_TOKEN=$$(grep "^GITHUB_TOKEN=" dev/compose.env | cut -d'=' -f2- | tr -d '"' | tr -d "'"); \
		if [ -z "$$GITHUB_TOKEN" ]; then \
			echo "Error: GITHUB_TOKEN not found in dev/compose.env"; \
			echo "Please add GITHUB_TOKEN=your_token to dev/compose.env"; \
			exit 1; \
		fi; \
	fi; \
	DOCKER_BUILDKIT=1 docker build \
		--build-arg GITHUB_TOKEN=$${GITHUB_TOKEN:-$$(grep "^GITHUB_TOKEN=" dev/compose.env 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'")} \
		--target prod \
		-f Dockerfile \
		-t ${NAMESPACE}_floxym:latest \
		.

dev-cert: ## Generates nginx SSL cert
	@mkdir -p dev/nginx/ssl
	@openssl req -newkey rsa:4096 -keyout dev/nginx/ssl/floxy.local.key -out dev/nginx/ssl/floxy.local.csr -nodes -subj "/C=RU/ST=Moscow/L=Moscow/O=Floxy/OU=Floxy/CN=floxy"
	@openssl x509 -req -in dev/nginx/ssl/floxy.local.csr -signkey dev/nginx/ssl/floxy.local.key -out dev/nginx/ssl/floxy.local.crt -days 365
