package config

import "maps"

type (
	Config struct {
		Urls      map[string]string
		Exchanges map[string]string
		Queue     map[string]string
	}
)


func NewConfig() *Config {
	config := &Config{
        Urls: make(map[string]string),
        Exchanges: make(map[string]string),
        Queue: make(map[string]string),
    }

	config.Exchanges["vote"] = "vote"
	config.Exchanges["poll"] = "poll"
	config.Exchanges["form"] = "form"
	config.Exchanges["answer"] = "answer"

	maps.Copy(config.Exchanges, config.Queue)

	config.Exchanges["self"] = "self"
	
    return config
}