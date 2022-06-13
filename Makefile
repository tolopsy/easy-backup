build:
	@echo "Building backup management tool..."
	@cd ./cmd/backup; \
	go build -o backup
	@echo "Management tool build complete!"

	@echo "Building backup daemon tool..."
	@cd ./cmd/backupd; \
	go build -o backupd
	@echo "Daemon buld complete!"
	