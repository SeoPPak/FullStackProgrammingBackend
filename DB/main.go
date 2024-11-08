package main

func main() {
	client, _, _ := ConnectDB()
	defer DisconnectDB(client)

	collection := SelectTable(client)
	InsertDocument(collection)
}
