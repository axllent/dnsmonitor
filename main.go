package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/axllent/ghru"
	"github.com/spf13/pflag"
)

var (
	customDNS   string
	interval    int
	help        bool
	showversion bool
	update      bool
	version     = "dev"
	configFile  string
	config      Config
)

// Domain struct
type Domain struct {
	Name       string
	LookupType string
	IPs        []string
}

func main() {

	flag := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	defaultConfig := HomeDir() + "/.config/dnsmonitor.json"

	flag.StringVarP(&configFile, "config", "c", defaultConfig, "config file")
	flag.StringVarP(&customDNS, "dns", "d", "", "custom dns server ip (defaults to system DNS)")
	flag.IntVarP(&interval, "interval", "i", 5, "interval to check in minutes")
	flag.BoolVarP(&help, "help", "h", false, "help")
	flag.MarkHidden("help")
	flag.BoolVarP(&showversion, "version", "v", false, "show version")
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

	flag.Parse(os.Args[1:])

	if help {
		flag.Usage()
		return
	}

	if showversion {
		fmt.Println(fmt.Sprintf("Version: %s", version))
		latest, _, _, err := ghru.Latest("axllent/dnsmonitor", "dnsmonitor")
		if err == nil && ghru.GreaterThan(latest, version) {
			fmt.Printf("Update available: %s\nRun `%s -u` to update.\n", latest, os.Args[0])
		}
		return
	}

	if update {
		rel, err := ghru.Update("axllent/dnsmonitor", "dnsmonitor", version)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Updated %s to version %s\n", os.Args[0], rel)
		return
	}

	configJSON, err := ioutil.ReadFile(configFile)
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

	domains := []*Domain{}

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
			log.Fatal(domain, "is not a valid domain")
		}

		domains = append(domains, &Domain{domain, lookupType, nil})
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Minute)

	for ; true; <-ticker.C {

		for _, domain := range domains {

			ips := Lookup(domain.LookupType, domain.Name)

			if domain.IPs == nil {
				// initial DNS resolution - no alerts
				log.Printf("Monitoring [%s] %s: %s\n", domain.LookupType, domain.Name, DNS2String(ips))

				title := fmt.Sprintf("Monitoring [%s] %s\n", domain.LookupType, domain.Name)
				message := fmt.Sprintf("status: %s", DNS2String(ips))

				SendNotifications(title, message, "1")
			} else if !Equal(domain.IPs, ips) {
				// do stuff
				log.Printf("[%s] %s UPDATED (%s)\n", domain.LookupType, domain.Name, DNS2String(ips))

				title := fmt.Sprintf("UPDATED [%s] %s\n", domain.LookupType, domain.Name)
				message := fmt.Sprintf("was: %s\nnow: %s", DNS2String(domain.IPs), DNS2String(ips))

				SendNotifications(title, message, "5")
			}

			domain.IPs = ips
		}
	}
}

// Lookup will do a DNS lookup and
// return an sorted array of results
func Lookup(lookupType, ip string) []string {
	r := net.Resolver{}
	if customDNS != "" {
		r.PreferGo = true
		r.Dial = CustomDialer
	}
	ctx := context.Background()

	result := []string{}

	if lookupType == "A" {
		ipaddr, err := r.LookupIPAddr(ctx, ip)
		if err == nil {
			for _, v := range ipaddr {
				result = append(result, fmt.Sprintf("%s", v.IP))
			}
		}
	} else if lookupType == "CNAME" {
		cname, err := r.LookupCNAME(ctx, ip)
		if err == nil {
			result = append(result, cname)
		}
	} else if lookupType == "MX" {
		ipaddr, err := r.LookupMX(ctx, ip)
		if err == nil {
			for _, v := range ipaddr {
				result = append(result, fmt.Sprintf("%s", v.Host))
			}
		}
	} else if lookupType == "TXT" {
		ipaddr, err := r.LookupTXT(ctx, ip)
		if err == nil {
			result = ipaddr
		}
	} else if lookupType == "NS" {
		ipaddr, err := r.LookupNS(ctx, ip)
		if err == nil {
			for _, v := range ipaddr {
				result = append(result, fmt.Sprintf("%s", v.Host))
			}
		}
	}

	// make sure they are in alphabetical order for Equal() comparison
	sort.Strings(result)

	return result
}

// DNS2String returns an IP slice as a string,
// or unresolved for return / output
func DNS2String(results []string) string {
	if len(results) == 0 {
		return "not found"
	}
	return strings.Join(results, ", ")
}

// Equal compares two slices
func Equal(a, b []string) bool {
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

// stringInSlice is a string-only version in php's in_array()
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// CustomDialer allows a custom DNS server to be set
func CustomDialer(ctx context.Context, network, address string) (net.Conn, error) {
	d := net.Dialer{}
	return d.DialContext(ctx, "udp", customDNS+":53")
}
