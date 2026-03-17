package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/viper"
)

// Account represents a single WebUntis account
type Account struct {
	Server            string `mapstructure:"server" json:"server"`
	School            string `mapstructure:"school" json:"school"`
	Username          string `mapstructure:"username" json:"username"`
	Password          string `mapstructure:"password" json:"password"`
	SessionID         string `mapstructure:"sessionid" json:"session_id"` // Classic session, may be needed for JWT exchange
	JWTToken          string `mapstructure:"jwttoken" json:"jwt_token"`
	TenantID          string `mapstructure:"tenantid" json:"tenant_id"`
	SchoolYearID      string `mapstructure:"schoolyearid" json:"school_year_id"`
	CSRFToken         string `mapstructure:"csrftoken" json:"csrf_token"`
	ActiveStudentID   int    `mapstructure:"activestudentid" json:"active_student_id"`
	ActiveStudentName string `mapstructure:"activestudentname" json:"active_student_name"`
}

// Config represents the persisted configuration
type Config struct {
	CurrentAccount string             `mapstructure:"current_account" json:"current_account"`
	Accounts       map[string]Account `mapstructure:"accounts" json:"accounts"`
}

func InitConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".webuntis-go-cli")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// Config file found and read
	}
}

// Helper to update viper state from struct and write
func writeConfig(c *Config) error {
	viper.Set("current_account", c.CurrentAccount)
	viper.Set("accounts", c.Accounts)

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, ".webuntis-go-cli.yaml")
	return viper.WriteConfigAs(configPath)
}

// AddOrUpdateAccount saves an account and optionally sets it as current
func AddOrUpdateAccount(name string, acc Account, makeCurrent bool) error {
	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		// Handle empty config case
		c = Config{Accounts: make(map[string]Account)}
	}
	if c.Accounts == nil {
		c.Accounts = make(map[string]Account)
	}

	c.Accounts[name] = acc
	if makeCurrent || c.CurrentAccount == "" {
		c.CurrentAccount = name
	}

	return writeConfig(&c)
}

// SwitchAccount changes the active account
func SwitchAccount(name string) error {
	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		return err
	}

	if _, ok := c.Accounts[name]; !ok {
		return fmt.Errorf("account '%s' not found", name)
	}

	c.CurrentAccount = name
	return writeConfig(&c)
}

// DeleteAccount removes an account
func DeleteAccount(name string) error {
	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		return err
	}

	if _, ok := c.Accounts[name]; !ok {
		return fmt.Errorf("account '%s' not found", name)
	}

	delete(c.Accounts, name)

	if c.CurrentAccount == name {
		c.CurrentAccount = ""
		// Pick first available if any
		for k := range c.Accounts {
			c.CurrentAccount = k
			break
		}
	}

	return writeConfig(&c)
}

// GetActiveAccount returns the configuration for the currently active account
func GetActiveAccount() (Account, string) {
	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		return Account{}, ""
	}

	if c.CurrentAccount == "" {
		return Account{}, ""
	}

	acc, ok := c.Accounts[c.CurrentAccount]
	if !ok {
		return Account{}, c.CurrentAccount
	}

	return acc, c.CurrentAccount
}

// ListAccounts returns a list of account names
func ListAccounts() []string {
	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		return []string{}
	}

	keys := make([]string, 0, len(c.Accounts))
	for k := range c.Accounts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// GetConfig returns the active account's data (compatibility function)
func GetConfig() Account {
	acc, _ := GetActiveAccount()
	return acc
}
