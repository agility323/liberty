package itf

type ServiceConfType struct {
	LogLevel int `json:"log_level"`
	Host int `json:"host"`
	ServiceType string `json:"service_type"`
	GateServerAddr string `json:"gate_server_addr"`
	ProfilePort int `json:"profile_port"`
	Etcd []string `json:"etcd"`
	ServiceRequestTimeout int `json:"service_request_timeout"`
	ServerHotfixPath string `json:"server_hotfix_path"`
}
