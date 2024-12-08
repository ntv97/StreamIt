package app

import (
        "encoding/json"
        "os"
)

// Config settings for main App.
type Config struct {
        Library []*PathConfig `json:"library"`
        Server  *ServerConfig `json:"server"`
}

// PathConfig settings for media library path.
type PathConfig struct {
        Path   string `json:"path"`
}

// ServerConfig settings for App Server.
type ServerConfig struct {
        Host string `json:"host"`
        Port int    `json:"port"`
}

// DefaultConfig returns Config initialized with default values.
func DefaultConfig() *Config {
        return &Config{
                Library: []*PathConfig{
                        &PathConfig{
                                Path:   "videos",
                                //Prefix: "",
                        },
                },
                Server: &ServerConfig{
                        Host: "127.0.0.1",
                        Port: 0,
                },
        }
}

// ReadFile reads a JSON file into Config.
func (c *Config) ReadFile(path string) error {
        f, err := os.Open(path)
        if err != nil {
                return err
        }
        defer f.Close()
        d := json.NewDecoder(f)
        return d.Decode(c)
}
