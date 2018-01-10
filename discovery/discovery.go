package discovery

// Subscriber defines interface for subscriber on nsqlookupds changes.
type Subscriber interface {
	DisconnectFromNSQLookupd(addr string) error
	ConnectToNSQLookupd(addr string) error
}
