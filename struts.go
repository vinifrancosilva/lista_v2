package main

type AppConfig struct {
	DbUser           string
	DbPassword       string
	DbHost           string
	DbPort           string
	DbName           string
	DbSSLMode        string
	DbSearchPath     string
	SessionSecretKey string
}

func (a *AppConfig) SessionSecretKeyByte() []byte {
	return []byte(a.SessionSecretKey)
}
