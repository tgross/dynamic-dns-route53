package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/miekg/dns"
)

const defaultTarget = "myip.opendns.com"
const defaultServer = "resolver1.opendns.com"
const defaultCachePath = "/tmp/.dynamic-dns-route53.cache"

type Config struct {
	target     string
	server     string
	cachePath  string
	zoneId     string
	domainName string
	ttl        int64
}

func NewConfig() *Config {
	target := flag.String("target", defaultTarget, "lookup target")
	server := flag.String("server", defaultServer, "resolver")
	cachePath := flag.String("path", defaultCachePath, "path to cache file")
	ttl := flag.Int64("ttl", 60, "TTL in seconds")
	zoneId := flag.String("zoneId", "", "Route53 zone ID")
	domainName := flag.String("name", "", "domain name")
	flag.Parse()

	if *zoneId == "" {
		log.Fatal("-zoneId flag must be set")
	}
	if *domainName == "" {
		log.Fatal("-name flag must be set")
	}
	return &Config{
		target:     *target,
		server:     *server,
		cachePath:  *cachePath,
		zoneId:     *zoneId,
		domainName: *domainName,
		ttl:        *ttl,
	}
}

func getIp(target, server string) (*dns.A, error) {
	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(target+".", dns.TypeA)
	r, _, err := c.Exchange(&m, server+":53")
	if err != nil {
		return nil, err
	}
	if len(r.Answer) == 0 {
		return nil, fmt.Errorf("no results")
	}
	for _, ans := range r.Answer {
		Arecord := ans.(*dns.A)
		return Arecord, nil
	}
	// note: should be unreachable
	return nil, fmt.Errorf("no results")
}

func isIPChanged(path string, ip *dns.A) bool {
	if _, err := os.Stat(path); err == nil {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		if string(content) == ip.String() {
			return true
		}
	} else if !os.IsNotExist(err) {
		log.Fatal(err)
	}
	content := []byte(ip.String())
	err := ioutil.WriteFile(path, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return false
}

func updateRecord(config *Config, ip *dns.A) error {
	zoneId := config.zoneId
	domainName := config.domainName
	ttl := config.ttl

	updateTime := time.Now().UTC().Format(time.RFC3339)
	updateMsg := fmt.Sprintf("update '%s' to %s at %s",
		domainName, ip.A, updateTime)

	awsSession := session.Must(session.NewSession())
	r53 := route53.New(awsSession)
	change := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(route53.ChangeActionUpsert),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(domainName),
						Type: aws.String(route53.RRTypeA),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(ip.String()),
							},
						},
						TTL:           aws.Int64(ttl),
						SetIdentifier: aws.String(updateMsg),
					},
				},
			},
			Comment: aws.String("Record managed by dynamic-dns-route53"),
		},
		HostedZoneId: aws.String(zoneId),
	}
	log.Print(updateMsg)
	_, err := r53.ChangeResourceRecordSets(change)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	log.SetFlags(log.Lshortfile)
	config := NewConfig()
	ip, err := getIp(config.target, config.server)
	if err != nil {
		log.Fatal(err)
	}
	if !isIPChanged(config.cachePath, ip) {
		return // no change to make
	}
	err = updateRecord(config, ip)
	if err != nil {
		log.Fatal(err)
	}
}
