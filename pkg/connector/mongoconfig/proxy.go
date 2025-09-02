package mongoconfig

type MongoProxy struct {
	Host string
	Port int
	User string
	Pass string
}

func (p *MongoProxy) Enabled() bool {
	return p.Host != "" && p.Port != 0 && p.User != "" && p.Pass != ""
}
