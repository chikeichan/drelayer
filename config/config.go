package config

type Config struct {
	DDRP         DDRP     `required:"true"`
	Database     Database `required:"true"`
	Server       Server   `required:"true"`
	TLDs         []TLD    `required:"true"`
	FeatureFlags FeatureFlags
}

type DDRP struct {
	Address string `required:"true"`
}

type Database struct {
	Debug       bool   `default:"false"`
	Name        string `required:"true"`
	Username    string `required:"true"`
	Password    string
	Host        string `required:"true"`
	Port        int    `default:"5432"`
	SSLMode     string `default:"verify-full"`
	SSLRootCert string
}

type TLD struct {
	Name       string
	PrivateKey string
}

type Server struct {
	Port       int    `default:"8080"`
	ServiceKey string `required:"true"`
}

type FeatureFlags struct {
	AllowSignup bool `default:"false"`
}
