run:
	go run .\cmd\banner\main.go --env=.\env\local.env
up:
	goose -dir migrations postgres postgres://postgres:183461@localhost:5432/postgres up
down:
	goose -dir migrations postgres postgres://postgres:183461@localhost:5432/postgres down