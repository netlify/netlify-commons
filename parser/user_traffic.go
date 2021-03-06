package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

//UserTraffic is a single decoded user traffic log line
type UserTraffic struct {
	Status       int       `json:"status"`
	RequestSize  int64     `json:"request_size"`
	ResponseSize int64     `json:"response_size"`
	Timing       int64     `json:"timing"`
	Timestamp    time.Time `json:"timestamp"`
	RequestID    string    `json:"request_id"`
	Result       string    `json:"result"`
	CSID         string    `json:"csid"`
	CCID         string    `jsond:"ccid"`
	CID          string    `json:"cid"`
	Proto        string    `json:"proto"`
	Method       string    `json:"method"`
	URL          string    `json:"url"`
	SID          string    `json:"sid"`
	AID          string    `json:"aid"`
	DID          string    `json:"did"`
	Cancel       string    `json:"cancel"`
	CCancel      string    `json:"ccancel"`
	ProxyType    string    `json:"proxy_type"`
	FID          string    `json:"fid"`
	ContentType  string    `json:"content_type"`
	Address      string    `json:"address"`
	Country      string    `json:"country"`
	Referrer     string    `json:"referrer"`
	CW           string    `json:"cw"`
	SSLVersion   string    `json:"ssl_version"`
	SSLCipher    string    `json:"ssl_cipher"`
	ENC          string    `json:"enc"`
	UserAgent    string    `json:"ua"`
	Unparsed     []string  `json:"unparsed"`
}

//ParseUserTrafficRecord parses a raw user traffic log line into a UserTraffic struct
//A slice of any unknown kv pairs or unbalanced fields will be appended to the UserTraffic
//structs Unparsed field.
func ParseUserTrafficRecord(raw string) (*UserTraffic, error) {
	var ut UserTraffic
	var err error

	if count := strings.Count(raw, `@timestamp`); count > 1 {
		return nil, fmt.Errorf("%d @timestamp fields detected", count)
	}

	praw := strings.SplitN(raw, " ua=", 2)

	if len(praw) > 1 {
		ut.UserAgent = praw[1]
	}

	for _, field := range strings.Fields(praw[0]) {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 { // most commonly due to kv's with duplicate value fields
			ut.Unparsed = append(ut.Unparsed, field)
			continue
		}
		switch parts[0] {
		case "request_id":
			ut.RequestID = parts[1]
		case "@timestamp", "timestamp":
			tsFloat, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return nil, fmt.Errorf("malformed field (%s) value: %s", parts[0], parts[1])
			}
			ut.Timestamp = time.Unix(int64(tsFloat), 0)
		case "timing":
			if ut.Timing, err = strconv.ParseInt(parts[1], 10, 64); err != nil {
				return nil, fmt.Errorf("malformed field (%s) value: %s", parts[0], parts[1])
			}
		case "result":
			ut.Result = parts[1]
		case "csid":
			ut.CSID = strings.TrimSuffix(parts[1], ",")
		case "cid":
			ut.CID = strings.TrimSuffix(parts[1], ",")
		case "ccid":
			ut.CCID = strings.TrimSuffix(parts[1], ",")
		case "status":
			if ut.Status, err = strconv.Atoi(parts[1]); err != nil {
				return nil, fmt.Errorf("malformed field (%s) value: %s", parts[0], parts[1])
			}
		case "request_size":
			if ut.RequestSize, err = strconv.ParseInt(parts[1], 10, 64); err != nil {
				return nil, fmt.Errorf("malformed field (%s) value: %s", parts[0], parts[1])
			}
		case "response_size":
			if ut.ResponseSize, err = strconv.ParseInt(parts[1], 10, 64); err != nil {
				return nil, fmt.Errorf("malformed field (%s) value: %s", parts[0], parts[1])
			}
		case "proto":
			ut.Proto = parts[1]
		case "method":
			ut.Method = parts[1]
		case "url":
			ut.URL = parts[1]
		case "sid":
			//SID's & AID's somewhat frequently have a trailing comma, while we try not to manipulate
			//or clean up the log lines inline this was an easy one that we felt we should
			//proactively handle
			ut.SID = strings.TrimSuffix(parts[1], ",")
		case "aid":
			//SID's & AID's somewhat frequently have a trailing comma, while we try not to manipulate
			//or clean up the log lines inline this was an easy one that we felt we should
			//proactively handle
			ut.AID = strings.TrimSuffix(parts[1], ",")
		case "did":
			ut.DID = strings.TrimSuffix(parts[1], ",")
		case "cancel":
			ut.Cancel = strings.TrimSuffix(parts[1], ",")
		case "proxy_type":
			ut.ProxyType = strings.TrimSuffix(parts[1], ",")
		case "fid":
			ut.FID = strings.TrimSuffix(parts[1], ",")
		case "content_type":
			ut.ContentType = strings.TrimSuffix(parts[1], ",")
		case "address":
			ut.Address = strings.TrimSuffix(parts[1], ",")
		case "country":
			ut.Country = strings.TrimSuffix(parts[1], ",")
		case "referrer":
			ut.Referrer = strings.TrimSuffix(parts[1], ",")
		case "cw":
			ut.CW = strings.TrimSuffix(parts[1], ",")
		case "ssl_version":
			ut.SSLVersion = strings.TrimSuffix(parts[1], ",")
		case "ssl_cipher":
			ut.SSLCipher = strings.TrimSuffix(parts[1], ",")
		case "enc":
			ut.ENC = strings.TrimSuffix(parts[1], ",")
		default:
			ut.Unparsed = append(ut.Unparsed, field)
		}
	}
	return &ut, nil
}
