package networking

type Contact struct { //the contacts list from this point.
	connectionString string //ip and port for the specified endpoint.
	NodeID           int
	//any other data needed about an endpoint would be stored here.
}
