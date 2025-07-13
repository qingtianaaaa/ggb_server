package config

import "fmt"

func IsProd() bool {
	return Cfg.Server.Mode == "prod"
}

func IsDev() bool {
	return Cfg.Server.Mode == "dev"
}

func IsTest() bool {
	return Cfg.Server.Mode == "test"
}

func (d *Database) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Name)
}
