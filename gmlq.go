package main

import "fmt"
import "os"
import "log"
import "gopkg.in/ldap.v2"

// todo
//  - config file
//	-- mail attr
//	-- search attrs
//	   display attrs
//  - objectify configuration?

type searchConfig struct {
	SearchAttributes []string
	LdapServer       string
	SearchBase       string
}

func defaultConfig() *searchConfig {
	return &searchConfig{
		SearchAttributes: []string{"uid", "cn"},
		LdapServer:       "ldap.corp.redhat.com",
		SearchBase:       "dc=redhat,dc=com",
	}
}

func getConfig() (*searchConfig, error) {
	// todo: read from a config file
	return defaultConfig(), nil
}

type searchResult struct {
	Mail  string
	Name  string
	Title string
}

func searchLdap(conf *searchConfig, term string) ([]searchResult, error) {
	// combine filter
	filter := "(&(|"
	for _, attr := range conf.SearchAttributes {
		filter = fmt.Sprintf("%s(%s=*%s*)", filter, attr, term)
	}
	filter = fmt.Sprintf("%s)(mail=*))", filter)

	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", conf.LdapServer, 389))
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		conf.SearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"mail", "cn", "rhatJobTitle"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	res := make([]searchResult, len(sr.Entries))
	for idx, entry := range sr.Entries {
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
