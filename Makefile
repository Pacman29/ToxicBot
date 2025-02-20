
### билдит докер образ для выравнивания структур
align-build:
	DOCKER_SCAN_SUGGEST=false docker build -t golang-align-check --no-cache - < $(PWD)/tools/align.Dockerfile

### выравнивает структуры для меньшей аллокации
align:
	docker run --rm -v $(PWD):/app -w /app golang-align-check fieldalignment -fix ./...

### врубает линтер
lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.50.1 golangci-lint run -v

### выравнивает импорты
imports:
	docker run --rm -v $(pwd):/data cytopia/goimports -d .