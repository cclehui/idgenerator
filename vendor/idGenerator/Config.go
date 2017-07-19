package idGenerator

type Config struct {
    Addr string `toml: "addr"`
    LogPath string `toml: "log_path"`
    LogLevel string `toml: "log_level"`
    Mysql Mysql `toml: "mysql"`
}

type Mysql struct {
    Host string `toml: "host"`
    Hort string `toml: "port"`
    Name string `toml: "name"`
    User string `toml: "user"`
    Password string `toml: "password"`
    MaxIdleConns string `toml: "max_idle_conns"`
}
