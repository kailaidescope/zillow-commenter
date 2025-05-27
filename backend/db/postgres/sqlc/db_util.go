package postgres

//// Using [golang-migrate](https://github.com/golang-migrate/migrate)

// To run migrate commands (from backend folder):
// migrate -path db/migrations -database "<connection_string>" -verbose <command_to_be_executed>

// To create a new migration (from backend folder):
// migrate create -ext sql -dir db/migrations -seq <name_of_migration>

//sqlc generate
//after having modified the query and schema files

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// returns a queries struct that can be used to execute queries and a
// function to close the connection linked to it
func GetConnection() (*pgx.Conn, error) {
	godotenv.Load()
	connStr := os.Getenv("CONNECTION_STRING")
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// use tx.Commit() if there are no errors at the end of the transaction
// if error call tx.Rollback()
// tx.Rollback() can be defered since if Commit is called first then rollback has no effect

// queries := New(conn)

// tx, err := conn.Begin(ctx)
