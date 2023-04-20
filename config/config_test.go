package config_test

import (
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	config "github.com/companieshouse/chs-streaming-api-frontend/config"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/chs.go/log/properties"
	. "github.com/smartystreets/goconvey/convey"
)

// key constants
const (
	BINDADDRESSKEYCONST             = `BIND_ADDRESS`
	CERTFILEKEYCONST                = `CERT_FILE`
	KEYFILEKEYCONST                 = `KEY_FILE`
	ERICLOCALURLKEYCONST            = `ERIC_LOCAL_URL`
	LOGTOPICKEYCONST                = `LOG_TOPIC`
	REQUESTTIMEOUTKEYCONST          = `REQUEST_TIMEOUT`
	HEARTBEATINTERVALKEYCONST       = `HEARTBEAT_INTERVAL`
	CACHEBROKERURLKEYCONST          = `CACHE_BROKER_URL`
	OFFICERSENDPOINTENABLEDKEYCONST = `OFFICERS_ENDPOINT_ENABLED`
	PSCSENDPOINTENABLEDKEYCONST     = `PSCS_ENDPOINT_ENABLED`
	CHSAPIKEYKEYCONST               = `CHS_API_KEY`
)

// value constants
const (
	bindAddressConst       = `localhost:1234`
	certConst              = `cert-file-test`
	keyFileConst           = `key-file-test`
	ericLocalURLConst      = `eric-url-test`
	logTopicConst          = `filing-logs-test`
	requestTimeoutConst    = 10
	heartbeatIntervalConst = 20
	cacheBrokerURLConst    = `cache-broker-url-test`
	officersEndpointConst  = false
	pscEndpointConst       = false
	chsAPIKeyConst         = `chs-api-key-1234`
	nameConst              = `chs-log-test`
	namespaceConst         = `chs-namespace-test`
	apiKeyConst            = `ApiKey`
	certFileFieldConst     = `CertFile`
	keyFileFieldConst      = `KeyFile`
	EricURLConst           = `EricURL`
	CacheBrokerURLConst    = `CacheBrokerURL`
	configConst            = `config`
)

func TestConfig(t *testing.T) {
	t.Parallel()
	os.Clearenv()
	var (
		err           error
		configuration *config.Config
		envVars       = map[string]string{
			BINDADDRESSKEYCONST:             bindAddressConst,
			CERTFILEKEYCONST:                certConst,
			KEYFILEKEYCONST:                 keyFileConst,
			ERICLOCALURLKEYCONST:            ericLocalURLConst,
			LOGTOPICKEYCONST:                logTopicConst,
			REQUESTTIMEOUTKEYCONST:          strconv.Itoa(requestTimeoutConst),
			HEARTBEATINTERVALKEYCONST:       strconv.Itoa(heartbeatIntervalConst),
			CACHEBROKERURLKEYCONST:          cacheBrokerURLConst,
			OFFICERSENDPOINTENABLEDKEYCONST: strconv.FormatBool(officersEndpointConst),
			PSCSENDPOINTENABLEDKEYCONST:     strconv.FormatBool(pscEndpointConst),
			CHSAPIKEYKEYCONST:               chsAPIKeyConst,
		}
		builtConfig = config.Config{
			BindAddress:          bindAddressConst,
			CertFile:             certConst,
			KeyFile:              keyFileConst,
			EricURL:              ericLocalURLConst,
			FilingLogs:           logTopicConst,
			RequestTimeout:       requestTimeoutConst,
			HeartbeatInterval:    heartbeatIntervalConst,
			CacheBrokerURL:       cacheBrokerURLConst,
			OfficersEndpointFlag: officersEndpointConst,
			PSCsEndpointFlag:     pscEndpointConst,
			ApiKey:               chsAPIKeyConst,
		}
		apiKeyRegex         = regexp.MustCompile(apiKeyConst)
		certFileRegex       = regexp.MustCompile(certFileFieldConst)
		keyFileRegex        = regexp.MustCompile(keyFileFieldConst)
		ericURLRegex        = regexp.MustCompile(EricURLConst)
		cacheBrokerURLRegex = regexp.MustCompile(CacheBrokerURLConst)
	)

	// set test env variables
	for varName, varValue := range envVars {
		os.Setenv(varName, varValue)
		defer os.Unsetenv(varName)
	}

	Convey("Given an environment with no environment variables set", t, func() {

		Convey("Then configuration should be nil", func() {
			So(configuration, ShouldBeNil)
		})

		Convey("When the config values are retrieved", func() {

			Convey("Then there should be no error returned, and values are as expected", func() {
				configuration, err = config.Get()
				So(err, ShouldBeNil)
				So(configuration, ShouldResemble, &builtConfig)
			})

			Convey("The generated JSON string from configuration should not contain sensitive data", func() {
				formatter := log.GetFormatter()
				jsonByte := []byte(formatter.Format(
					map[string]interface{}{
						properties.CreatedLogKey:   time.Now(),
						properties.EventLogKey:     nameConst,
						properties.NamespaceLogKey: namespaceConst,
					}, map[string]interface{}{
						configConst: builtConfig,
					},
				))
				So(apiKeyRegex.Match(jsonByte), ShouldEqual, false)
				So(certFileRegex.Match(jsonByte), ShouldEqual, false)
				So(keyFileRegex.Match(jsonByte), ShouldEqual, false)
				So(ericURLRegex.Match(jsonByte), ShouldEqual, false)
				So(cacheBrokerURLRegex.Match(jsonByte), ShouldEqual, false)
			})
		})
	})
}
