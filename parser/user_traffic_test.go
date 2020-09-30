package parser

import (
	"bytes"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//raw record to test against as a fail-safe (incase the template drifts)
var rawUTRecord = "request_id=c9948493-1ece-4d21-a2d1-f96a9feded3c @timestamp=1585844380.949 timing=1 result=TCP_MEM_HIT cid=- ccid=12345 status=200 request_size=1 response_size=66000 proto=http/2 method=GET url=http://localhost/something/1591294965428966000/something.jpg sid=18bb190b-6727-497a-af8b-f03287d14caf, aid=1591294965428966000 did=5e85df2043933dd053ebec6f cancel=- proxy_type=- stuff=things oneother=\"onething\" fid=- content_type=text/plain address=2605:6000:1714:56e:c98a:445c:febd:6baf country=US referrer=localhost cw=- ssl_version=TLSv1.2 ssl_cipher=ECDHE-RSA-AES256-GCM-SHA384 enc=- ua=Mozilla/5.0 (X11; CrOS x86_64 12239.92.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.136 Safari/537.36"

var utLineTemplateStr = "request_id={{.requestIDField}} " +
	"@timestamp={{.atTimestampField}} " +
	"timing={{.timingField}} " +
	"result={{.resultField}} " +
	"cid={{.cidField}} " +
	"ccid={{.ccidField}} " +
	"status={{.statusField}} " +
	"request_size={{.requestSizeField}} " +
	"response_size={{.responseSizeField}} " +
	"proto={{.protoField}} " +
	"method={{.methodField}} " +
	"url={{.urlField}} " +
	"sid={{.sidField}} " +
	"aid={{.aidField}} " +
	"did={{.didField}} " +
	"cancel={{.cancelField}} " +
	"proxy_type={{.proxyTypeField}} " +
	"{{.extraFields}} " + //extra fields we should gracefully handle by adding to the others map
	"fid={{.fidField}} " +
	"content_type={{.contentTypeField}} " +
	"address={{.addressField}} " +
	"country={{.countryField}} " +
	"referrer={{.referrerField}} " +
	"cw={{.cwField}} " +
	"ssl_version={{.sslVersionField}} " +
	"ssl_cipher={{.sslCipherField}} " +
	"enc={{.encField}} " +
	"ua={{.uaField}}"

var utLineTemplate = template.Must(template.New("user_traffic").Parse(utLineTemplateStr))

var (
	extraFields       = "stuff=things oneother=\"onething\""
	requestIDField    = "c9948493-1ece-4d21-a2d1-f96a9feded3c"
	atTimestampField  = "1585844380.949"
	timingField       = "1"
	resultField       = "TCP_MEM_HIT"
	cidField          = "-"
	ccidField         = "12345"
	statusField       = "200"
	requestSizeField  = "1"
	responseSizeField = "66000"
	protoField        = "http/2"
	methodField       = "GET"
	urlField          = "http://localhost/something/1591294965428966000/something.jpg"
	sidField          = "18bb190b-6727-497a-af8b-f03287d14caf"
	aidField          = "1591294965428966000"
	didField          = "5e85df2043933dd053ebec6f"
	cancelField       = "-"
	proxyTypeField    = "-"
	fidField          = "-"
	contentTypeField  = "text/plain"
	addressField      = "2605:6000:1714:56e:c98a:445c:febd:6baf"
	countryField      = "US"
	referrerField     = "localhost"
	cwField           = "-"
	uaField           = "Mozilla/5.0 (X11; CrOS x86_64 12239.92.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.136 Safari/537.36"
	sslVersionField   = "TLSv1.2"
	sslCipherField    = "ECDHE-RSA-AES256-GCM-SHA384"
	encField          = "-"
)

func defaultValues() map[string]string {
	return map[string]string{
		"extraFields":       extraFields,
		"requestIDField":    requestIDField,
		"atTimestampField":  atTimestampField,
		"timingField":       timingField,
		"resultField":       resultField,
		"cidField":          cidField,
		"ccidField":         ccidField,
		"statusField":       statusField,
		"requestSizeField":  requestSizeField,
		"responseSizeField": responseSizeField,
		"protoField":        protoField,
		"methodField":       methodField,
		"urlField":          urlField,
		"sidField":          sidField,
		"aidField":          aidField,
		"didField":          didField,
		"cancelField":       cancelField,
		"proxyTypeField":    proxyTypeField,
		"fidField":          fidField,
		"contentTypeField":  contentTypeField,
		"addressField":      addressField,
		"countryField":      countryField,
		"referrerField":     referrerField,
		"cwField":           cwField,
		"uaField":           uaField,
		"sslVersionField":   sslVersionField,
		"sslCipherField":    sslCipherField,
		"encField":          encField,
	}
}

func genUserTrafficLine(t *testing.T, values map[string]string) string {
	buf := new(bytes.Buffer)
	require.NoError(t, utLineTemplate.Execute(buf, values))
	return buf.String()
}

func TestParseUserTrafficPayload(t *testing.T) {
	expected := &UserTraffic{
		Status:       200,
		RequestSize:  1,
		ResponseSize: 66000,
		Timing:       1,
		Timestamp:    time.Unix(int64(1585844380), 0),
		RequestID:    "c9948493-1ece-4d21-a2d1-f96a9feded3c",
		Result:       "TCP_MEM_HIT",
		CSID:         "",
		CID:          "-",
		CCID:         "12345",
		Proto:        "http/2",
		Method:       "GET",
		URL:          "http://localhost/something/1591294965428966000/something.jpg",
		SID:          "18bb190b-6727-497a-af8b-f03287d14caf",
		AID:          "1591294965428966000",
		DID:          "5e85df2043933dd053ebec6f",
		Cancel:       "-",
		CCancel:      "",
		ProxyType:    "-",
		FID:          "-",
		ContentType:  "text/plain",
		Address:      "2605:6000:1714:56e:c98a:445c:febd:6baf",
		Country:      "US",
		Referrer:     "localhost",
		SSLCipher:    "ECDHE-RSA-AES256-GCM-SHA384",
		SSLVersion:   "TLSv1.2",
		ENC:          "-",
		CW:           "-",
		UserAgent:    "Mozilla/5.0 (X11; CrOS x86_64 12239.92.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.136 Safari/537.36",
		Other:        map[string]string{"stuff": "things", "oneother": "\"onething\""},
	}

	ut, err := ParseUserTrafficRecord(genUserTrafficLine(t, defaultValues()))
	require.NoError(t, err)
	assert.Equal(t, expected, ut)

	// quick fail safe - we should *always* be able to snag a raw log line and parse out these key fields.
	failSafeUT := &UserTraffic{
		Status:       200,
		RequestSize:  1,
		ResponseSize: 66000,
		Timing:       1,
		Timestamp:    time.Unix(int64(1585844380), 0),
		RequestID:    "c9948493-1ece-4d21-a2d1-f96a9feded3c",
		URL:          "http://localhost/something/1591294965428966000/something.jpg",
		SID:          "18bb190b-6727-497a-af8b-f03287d14caf",
		AID:          "1591294965428966000",
		DID:          "5e85df2043933dd053ebec6f",
		Address:      "2605:6000:1714:56e:c98a:445c:febd:6baf",
	}

	ut, err = ParseUserTrafficRecord(rawUTRecord)
	require.NoError(t, err)
	assert.Equal(t, failSafeUT.Status, ut.Status)
	assert.Equal(t, failSafeUT.RequestSize, ut.RequestSize)
	assert.Equal(t, failSafeUT.ResponseSize, ut.ResponseSize)
	assert.Equal(t, failSafeUT.Timing, ut.Timing)
	assert.Equal(t, failSafeUT.Timestamp, ut.Timestamp)
	assert.Equal(t, failSafeUT.URL, ut.URL)
	assert.Equal(t, failSafeUT.SID, ut.SID)
	assert.Equal(t, failSafeUT.AID, ut.AID)
	assert.Equal(t, failSafeUT.DID, ut.DID)
	assert.Equal(t, failSafeUT.Address, ut.Address)
}

func TestSidWithComma(t *testing.T) {
	fields := defaultValues()
	fields["sidField"] = fields["sidField"] + ","
	ut, err := ParseUserTrafficRecord(genUserTrafficLine(t, fields))
	require.NoError(t, err)
	assert.Equal(t, sidField, ut.SID)
}

func TestAidWithComma(t *testing.T) {
	fields := defaultValues()
	fields["aidField"] = fields["aidField"] + ","
	ut, err := ParseUserTrafficRecord(genUserTrafficLine(t, fields))
	require.NoError(t, err)
	assert.Equal(t, sidField, ut.SID)
}

func TestErrOnExtraTimestamp(t *testing.T) {
	withExtraTimestamp := "@timestamp=181818181 " + genUserTrafficLine(t, defaultValues())
	ut, err := ParseUserTrafficRecord(withExtraTimestamp)
	require.Error(t, err)
	require.Nil(t, ut)
}

func TestErrOnKeyWithNoValue(t *testing.T) {
	keyMissingValue := "randomKeyWithNoValue " + genUserTrafficLine(t, defaultValues())
	ut, err := ParseUserTrafficRecord(keyMissingValue)
	require.Error(t, err)
	require.Nil(t, ut)
}

func TestMalformedTimestamp(t *testing.T) {
	fields := defaultValues()
	fields["atTimestampField"] = "time_is_relative"
	ut, err := ParseUserTrafficRecord(genUserTrafficLine(t, fields))
	require.Error(t, err)
	require.Nil(t, ut)
}

func TestMalformedTiming(t *testing.T) {
	fields := defaultValues()
	fields["timingField"] = "somestring"
	ut, err := ParseUserTrafficRecord(genUserTrafficLine(t, fields))
	require.Error(t, err)
	require.Nil(t, ut)
}

func TestMalformedStatus(t *testing.T) {
	fields := defaultValues()
	fields["statusField"] = "somestring"
	ut, err := ParseUserTrafficRecord(genUserTrafficLine(t, fields))
	require.Error(t, err)
	require.Nil(t, ut)
}

func TestMalformedRequestSize(t *testing.T) {
	fields := defaultValues()
	fields["requestSizeField"] = "somestring"
	ut, err := ParseUserTrafficRecord(genUserTrafficLine(t, fields))
	require.Error(t, err)
	require.Nil(t, ut)
}

func TestMalformedResponseSize(t *testing.T) {
	fields := defaultValues()
	fields["responseSizeField"] = "somestring"
	ut, err := ParseUserTrafficRecord(genUserTrafficLine(t, fields))
	require.Error(t, err)
	require.Nil(t, ut)
}
