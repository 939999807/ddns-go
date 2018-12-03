package main

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var cfg Config
var alidnsClient alidns.Client

func run() {

	resp, err := http.Get(cfg.GetIpUrl)
	if err != nil {
		log.Printf("getIp err %v", err)
	}
	if resp != nil {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("http body close err %v", err)
			}
		}()
	}
	body, err := ioutil.ReadAll(resp.Body)
	currentIp := string(body);
	log.Println("currentIp " + currentIp)
	for _, describeDomainRecord := range cfg.DescribeDomainRecords {
		var recordIp string
		var recordId string
		{
			request := alidns.CreateDescribeDomainRecordsRequest()
			request.DomainName = cfg.DomainName
			request.TypeKeyWord = describeDomainRecord.TypeKeyWord
			request.RRKeyWord = describeDomainRecord.ResourceRecord
			response, err := alidnsClient.DescribeDomainRecords(request)
			if err != nil {
				// Handle exceptions
				log.Printf("err %v", err)
			}
			records := response.DomainRecords.Record
			if len(records) > 0 {
				record := records[0]
				recordIp = record.Value
				recordId = record.RecordId
				log.Println("recordIP " + recordIp)
			}
		}
		if recordIp == "" {
			log.Println("record not found")
			continue
		}
		if currentIp != recordIp {
			request := alidns.CreateUpdateDomainRecordRequest()
			request.RecordId = recordId
			request.RR = describeDomainRecord.ResourceRecord
			request.Type = describeDomainRecord.TypeKeyWord
			request.Value = currentIp
			_, err := alidnsClient.UpdateDomainRecord(request)
			if err != nil {
				// Handle exceptions
				log.Printf("err %v", err)
			}
			log.Println("update " + currentIp)
		}
	}
}

func main() {
	cfg = parseConfig()

	// Create an ECS client
	client, err := alidns.NewClientWithAccessKey(
		cfg.RegionId,    // Your Region ID
		cfg.AccessKeyId, // Your AccessKey ID
		cfg.AccessKeySecret) // Your AccessKey Secret
	if err != nil {
		// Handle exceptions
		panic(err)
	}
	alidnsClient = *client
	run()
	tick := time.Tick(time.Second * time.Duration(cfg.Interval))
	for _ = range tick {
		run()
	}
}
