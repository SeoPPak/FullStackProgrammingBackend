package server

import (
	"server/db"
)

func main() {
	client, _, _ := db.ConnectDB()
	defer db.DisconnectDB(client)

	collection := db.SelectTable(client)
	r := Setup()
	r.Run(":5000")
}
