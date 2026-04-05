package config

func (c *Config) SetUser(name string) {
	c.Current_user_name = name

	writeToConfig(*c)

}
