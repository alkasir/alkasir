package db

// Clients
type Clients struct {
	Internet InternetClient
	DB       DBClient
	Maxmind  MaxmindClient
}
