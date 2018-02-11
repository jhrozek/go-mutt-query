package main

import "fmt"
import "os"
import "log"
import "github.com/spf13/viper"
import "gopkg.in/ldap.v2"

// todo
//  - config file
//	-- mail attr
//	-- search attrs
//	   display attrs
//  - CI
//  - vendor packages

type searchConfig struct {
	LdapServer string `mapstructure:"uri"`
	LdapPort   int    `mapstructure:"port"`
	SearchBase string `mapstructure:"search_base"`

	MatchAttributes []string `mapstructure:"search_attrs"`

	DisplayAttributes []string `mapstructure:"display_attrs"`
}

func configFileSources() {
	viper.SetConfigName("gmlq")        // name of config file (without extension)
	viper.AddConfigPath("$HOME/.gmlq") // call multiple times to add many search paths
	viper.AddConfigPath(".")           // optionally look for config in the working directory
}

func configDefaults() {
	// set the defaults. Can be done for some attributes only
	viper.SetDefault("MatchAttributes", []string{"uid", "cn"})
	viper.SetDefault("DisplayAttributes", []string{"mail", "cn"})
	viper.SetDefault("LdapPort", 389)
}

func configSetUp() {
	configFileSources()
	configDefaults()
}

func getConfig() (*searchConfig, error) {
	configSetUp()

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		return nil, fmt.Errorf("Fatal error opening the config file: %s \n", err)
	}

	var sc searchConfig

	err = viper.Unmarshal(&sc)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}
	// todo: read from a config file
	return &sc, nil
}

type searchResult struct {
	Mail  string
	Name  string
	Title string
}

func searchLdap(conf *searchConfig, term string) ([]searchResult, error) {
	// combine filter
	filter := "(&(|"
	for _, attr := range conf.MatchAttributes {
		filter = fmt.Sprintf("%s(%s=*%s*)", filter, attr, term)
	}
	filter = fmt.Sprintf("%s)(mail=*))", filter)

	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", conf.LdapServer, conf.LdapPort))
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		conf.SearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		conf.DisplayAttributes,
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	res := make([]searchResult, len(sr.Entries))
	for idx, entry := range sr.Entries {
		// build a dict based on what the user configured to be in display attrs
		res[idx].Mail = entry.GetAttributeValue("mail")
		res[idx].Name = entry.GetAttributeValue("cn")
		res[idx].Title = entry.GetAttributeValue("rhatJobTitle")
	}
	return res, nil
}

func printResult(ldapRes []searchResult) {
	for _, entry := range ldapRes {
		fmt.Printf("%s\t%s\t%s\n", entry.Mail, entry.Name, entry.Title)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: gmlq TERM")
		os.Exit(0)
	}

	// todo: handle error
	c, err := getConfig()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	ldapRes, err := searchLdap(c, os.Args[1])
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	printResult(ldapRes)
}
