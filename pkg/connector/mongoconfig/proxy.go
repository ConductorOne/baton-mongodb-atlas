package mongoconfig

type MongoProxy struct {
	Host string
	Port int
}

func (p *MongoProxy) Enabled() bool {
	return p.Host != "" && p.Port != 0
}
