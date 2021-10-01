package node

// ReponseInfo holds base contact information for a node that may not be stored inside the courier service
type ResponseInfo struct {
	Id      string
	Address string
	Port    string
}
