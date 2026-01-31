package resume

import "strings"

type DBProtocol string

const (
	ProtocolMySQL      DBProtocol = "mysql"
	ProtocolPostgreSQL DBProtocol = "postgresql"
)

func getDBProtocol(protocol string) DBProtocol {
	switch strings.ToLower(protocol) {
	case "mysql":
		return ProtocolMySQL
	case "postgresql", "postgres":
		return ProtocolPostgreSQL
	default:
		return ProtocolPostgreSQL
	}
}
