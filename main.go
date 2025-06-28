// Package main is the main application
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/axllent/ghru/v2"
	"github.com/spf13/pflag"
)

var (
	customDNS   string
	interval    int
	help        bool
	showVersion bool
	update      bool
	version     = "dev"
	configFile  string
	config      configStruct

	ghruConf = ghru.Config{
		Repo:           "axllent/dnsmonitor",
		ArchiveName:    "dnsmonitor-{{.OS}}-{{.Arch}}",
		BinaryName:     "dnsmonitor",
		CurrentVersion: version,
	}
)

// Domain struct
type domainStruct struct {
	Name       string
	LookupType string
	IPs        []string
}

func main() {

	flag := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	defaultConfig := homeDir() + "/.config/dnsmonitor.json"

	flag.StringVarP(&configFile, "config", "c", defaultConfig, "config file")
	flag.StringVarP(&customDNS, "dns", "d", "", "custom dns server ip (defaults to system DNS)")
	flag.IntVarP(&interval, "interval", "i", 5, "interval to check in minutes")
	flag.BoolVarP(&help, "help", "h", false, "help")
	_ = flag.MarkHidden("help")
	flag.BoolVarP(&showVersion, "version", "v", false, "show version")
	flag.BoolVarP(&update, "update", "u", false, "update to latest version")

	flag.Usage = func() {
		fmt.Printf("DNS Monitor - A simple DNS monitor to alert DNS changes.\n\n")
		fmt.Println("Usage examples:")
		fmt.Printf("  %s www.example.com\n", os.Args[0])
		fmt.Printf("  %s mx:example.com\n", os.Args[0])
		fmt.Printf("  %s ns:example.com txt:example.com www.example.com\n", os.Args[0])
		fmt.Printf("  %s -d 1.1.1.1 example.com\n\n", os.Args[0])
		fmt.Printf("Valid query types are: a, cname, mx, txt & ns. The default is \"a\".\n\n")
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	if err := flag.Parse(os.Args[1:]); err != nil {
		fmt.Println("Error parsing flags:", err)
		os.Exit(1)
	}

	if help {
		flag.Usage()
		return
	}

	if showVersion {
		fmt.Printf("Version: %s\n", version)

		release, err := ghruConf.Latest()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// The latest version is the same version
		if release.Tag == version {
			os.Exit(0)
		}

		// A newer release is available
		fmt.Printf(
			"Update available: %s\nRun `%s -u` to update (requires read/write access to install directory).\n",
			release.Tag,
			os.Args[0],
		)
		os.Exit(0)
	}

	if update {
		// Update the app
		rel, err := ghruConf.SelfUpdate()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Updated %s to version %s\n", os.Args[0], rel.Tag)
		os.Exit(0)
	}

	configJSON, err := os.ReadFile(configFile)
	if err == nil {
		err = json.Unmarshal(configJSON, &config)

		if err != nil {
			fmt.Printf("Error parsing %s:\n\n%s\n\n", configFile, err)
			os.Exit(2)
		}
	}

	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	domains := []*domainStruct{}

	// regex for type
	reType := regexp.MustCompile(`^(a|mx|cname|txt|ns)\:([a-z0-9\.\-]{4,})$`)

	// quick & dirty domain validation
	reDomain := regexp.MustCompile(`^([a-z0-9\.\-]{4,})$`)

	for _, domain := range args {
		// set default lookup type
		lookupType := "A"
		domain = strings.ToLower(domain)

		matches := reType.FindStringSubmatch(domain)
		if len(matches) == 3 {
			lookupType = strings.ToUpper(matches[1])
			domain = matches[2]
		}

		if !reDomain.MatchString(domain) {
			fmt.Println("Not a valid domain:", domain)
			os.Exit(1)
		}

		domains = append(domains, &domainStruct{domain, lookupType, nil})
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Minute)

	for ; true; <-ticker.C {

		for _, domain := range domains {

			ips := lookup(domain.LookupType, domain.Name)

			if domain.IPs == nil {
				// initial DNS resolution
				log.Printf("Monitoring [%s] %s: %s\n", domain.LookupType, domain.Name, dnsToString(ips))

				title := fmt.Sprintf("Monitoring [%s] %s\n", domain.LookupType, domain.Name)
				message := fmt.Sprintf("status: %s", dnsToString(ips))

				sendNotifications(title, message, "1")
			} else if !equal(domain.IPs, ips) {
				// do stuff
				log.Printf("[%s] %s UPDATED (%s)\n", domain.LookupType, domain.Name, dnsToString(ips))

				title := fmt.Sprintf("UPDATED [%s] %s\n", domain.LookupType, domain.Name)
				message := fmt.Sprintf("was: %s\nnow: %s", dnsToString(domain.IPs), dnsToString(ips))

				sendNotifications(title, message, "5")
			}

			domain.IPs = ips
		}
	}
}

// Lookup will do a DNS lookup and
// return an sorted array of results
func lookup(lookupType, ip string) []string {
	r := net.Resolver{}
	if customDNS != "" {
		r.PreferGo = true
		r.Dial = customDialer
	}
	ctx := context.Background()

	result := []string{}

	switch lookupType {
	case "A":
		ipaddr, err := r.LookupIPAddr(ctx, ip)
		if err == nil {
			for _, v := range ipaddr {
				result = append(result, v.IP.String())
			}
		}
	case "CNAME":
		cname, err := r.LookupCNAME(ctx, ip)
		if err == nil {
			result = append(result, cname)
		}
	case "MX":
		ipaddr, err := r.LookupMX(ctx, ip)
		if err == nil {
			for _, v := range ipaddr {
				result = append(result, v.Host)
			}
		}
	case "TXT":
		ipaddr, err := r.LookupTXT(ctx, ip)
		if err == nil {
			result = ipaddr
		}
	case "NS":
		ipaddr, err := r.LookupNS(ctx, ip)
		if err == nil {
			for _, v := range ipaddr {
				result = append(result, string(v.Host))
			}
		}
	}

	// make sure they are in alphabetical order for Equal() comparison
	sort.Strings(result)

	return result
}

// dnsToString returns an IP slice as a string,
// or unresolved for return / output
func dnsToString(results []string) string {
	if len(results) == 0 {
		return "not found"
	}
	return strings.Join(results, ", ")
}

// Equal compares two slices
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// CustomDialer allows a custom DNS server to be set
func customDialer(ctx context.Context, _, _ string) (net.Conn, error) {
	d := net.Dialer{}
	return d.DialContext(ctx, "udp", customDNS+":53")
}
