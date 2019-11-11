package echo

type Writer interface {
	write(*event) error
}

type ConfigI interface {
	Init() Writer
}

type MonGoConfig struct {
	Url string
}

type ESConfig struct {
	Url      string
	Index    string
	Scheme   string
	Username string
	Password string
	Close    bool
}

func (EC *ESConfig) Init() Writer {
	if EC.Index == "" {
		EC.Index = "iot"
	}
	if EC.Scheme == "" {
		EC.Scheme = "http"
	}
	if EC.Url == "" {
		EC.Url = "http://192.168.1.6:9200"
	}
	return NewESClients(EC)
}
