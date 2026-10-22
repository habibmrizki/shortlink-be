include ./.env
DBURL=postgres://$(DBUSER):$(DBPASS)@$(DBHOST):$(DBPORT)/$(DBNAME)?sslmode=disable
MIGRATIONPATH=db/migrations
SEEDPATH=db/seeds

.PHONY: migrate-create migrate-createUp migrate-createDown migrate-status migrate-force seed seeds

migrate-create:
	@test -n "$(NAME)" || (echo "Usage: make migrate-create NAME=initial"; exit 1)
	@migrate create -ext sql -dir $(MIGRATIONPATH) -seq create_$(NAME)_table

migrate-createUp:
	@migrate -database $(DBURL) -path $(MIGRATIONPATH) up

migrate-createDown:
	@migrate -database $(DBURL) -path $(MIGRATIONPATH) down $(s)

migrate-status:
	@migrate -database $(DBURL) -path $(MIGRATIONPATH) version

migrate-force:
	@migrate -database $(DBURL) -path $(MIGRATIONPATH) force $(v)

seed seeds:
	@test -s "$(SEEDPATH)" || (echo "Seed file is empty: $(SEEDPATH)"; exit 1)
	@echo "Applying seed file: $(SEEDPATH)"
	@PGPASSWORD=$(DBPASS) psql -v ON_ERROR_STOP=1 -h $(DBHOST) -p $(DBPORT) -U $(DBUSER) -d $(DBNAME) -f $(SEEDPATH)
	@echo "Seed completed successfully."