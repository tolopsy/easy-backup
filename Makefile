build:
	@echo "Building backup management tool..."
	go build -o dist/ ./cmd/backup
	@echo "Management tool build complete!"

	@echo "Building backup daemon tool..."
	go build -o dist/ ./cmd/backupd
	@echo "Daemon buld complete!"
	