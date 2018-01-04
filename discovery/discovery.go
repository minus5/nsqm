package discovery

type Subscriber interface {
	DisconnectFromNSQLookupd(addr string) error
	ConnectToNSQLookupd(addr string) error
}
