package posthog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2"
)

const (
	unimplementedError = "not implemented"
	CACHE_DEFAULT_SIZE = 300_000

	propertyGeoipDisable = "$geoip_disable"
)

// Client interface is the main API exposed by the posthog package.
// Values that satisfy this interface are returned by the client constructors
// provided by the package and provide a way to send messages via the HTTP API.
type Client interface {
	io.Closer

	// Enqueue queues a message to be sent by the client when the conditions for a batch
	// upload are met.
	// This is the main method you'll be using, a typical flow would look like
	// this:
	//
	//	client := posthog.New(apiKey)
	//	...
	//	client.Enqueue(posthog.Capture{ ... })
	//	...
	//	client.Close()
	//
	// The method returns an error if the message queue could not be queued, which
	// happens if the client was already closed at the time the method was
	// called or if the message was malformed.
	Enqueue(Message) error

	// IsFeatureEnabled returns if a feature flag is on for a given user based on their distinct ID
	IsFeatureEnabled(FeatureFlagPayload) (interface{}, error)

	// GetFeatureFlag returns variant value if multivariant flag or otherwise a boolean indicating
	// if the given flag is on or off for the user
	GetFeatureFlag(FeatureFlagPayload) (interface{}, error)

	// GetFeatureFlagPayload returns feature flag's payload value matching key for user (supports multivariate flags).
	GetFeatureFlagPayload(FeatureFlagPayload) (string, error)

	// GetRemoteConfigPayload returns decrypted feature flag payload value for remote config flags.
	GetRemoteConfigPayload(string) (string, error)

	// GetAllFlags returns all flags for a user
	GetAllFlags(FeatureFlagPayloadNoKey) (map[string]interface{}, error)

	// ReloadFeatureFlags forces a reload of feature flags
	// NB: This is only available when using a PersonalApiKey
	ReloadFeatureFlags() error

	// GetFeatureFlags gets all feature flags, for testing only.
	// NB: This is only available when using a PersonalApiKey
	GetFeatureFlags() ([]FeatureFlag, error)

	// GetLastCapturedEvent returns the last captured event
	GetLastCapturedEvent() *Capture
}

type client struct {
	Config
	key string

	// This channel is where the `Enqueue` method writes messages so they can be
	// picked up and pushed by the backend goroutine taking care of applying the
	// batching rules.
	msgs chan APIMessage

	// These two channels are used to synchronize the client shutting down when
	// `Close` is called.
	// The first channel is closed to signal the backend goroutine that it has
	// to stop, then the second one is closed by the backend goroutine to signal
	// that it has finished flushing all queued messages.
	quit     chan struct{}
	shutdown chan struct{}

	// This HTTP client is used to send requests to the backend, it uses the
	// HTTP transport provided in the configuration.
	http http.Client

	// A background poller for fetching feature flags
	featureFlagsPoller *FeatureFlagsPoller

	distinctIdsFeatureFlagsReported *lru.Cache[flagUser, struct{}]

	// Last captured event
	lastCapturedEvent *Capture
	// Mutex to protect last captured event
	lastEventMutex sync.RWMutex

	// Decider for feature flag methods
	decider decider
}

type flagUser struct {
	distinctID string
	flagKey    string
}

// Instantiate a new client that uses the write key passed as first argument to
// send messages to the backend.
// The client is created with the default configuration.
func New(apiKey string) Client {
	// Here we can ignore the error because the default config is always valid.
	c, _ := NewWithConfig(apiKey, Config{})
	return c
}

// Instantiate a new client that uses the write key and configuration passed as
// arguments to send messages to the backend.
// The function will return an error if the configuration contained impossible
// values (like a negative flush interval for example).
// When the function returns an error the returned client will always be nil.
func NewWithConfig(apiKey string, config Config) (cli Client, err error) {
	if err = config.validate(); err != nil {
		return
	}

	config = makeConfig(config)
	reportedCache, err := lru.New[flagUser, struct{}](CACHE_DEFAULT_SIZE)
	if err != nil && config.Logger != nil {
		config.Logger.Errorf("Error creating cache for reported flags: %v", err)
	}
	c := &client{
		Config:                          config,
		key:                             apiKey,
		msgs:                            make(chan APIMessage, 100),
		quit:                            make(chan struct{}),
		shutdown:                        make(chan struct{}),
		http:                            makeHttpClient(config.Transport),
		distinctIdsFeatureFlagsReported: reportedCache,
	}

	c.decider = newFlagsClient(apiKey, config.Endpoint, c.http, config.FeatureFlagRequestTimeout, c.Errorf)

	if len(c.PersonalApiKey) > 0 {
		c.featureFlagsPoller = newFeatureFlagsPoller(
			c.key,
			c.Config.PersonalApiKey,
			c.Errorf,
			c.Endpoint,
			c.http,
			c.DefaultFeatureFlagsPollingInterval,
			c.NextFeatureFlagsPollingTick,
			c.FeatureFlagRequestTimeout,
			c.decider,
			c.Config.GetDisableGeoIP(),
		)
	}

	go c.loop()

	cli = c
	return
}

func makeHttpClient(transport http.RoundTripper) http.Client {
	httpClient := http.Client{
		Transport: transport,
	}
	if supportsTimeout(transport) {
		httpClient.Timeout = 10 * time.Second
	}
	return httpClient
}

func dereferenceMessage(msg Message) Message {
	switch m := msg.(type) {
	case *Alias:
		if m == nil {
			return nil
		}
		return *m
	case *Identify:
		if m == nil {
			return nil
		}
		return *m
	case *GroupIdentify:
		if m == nil {
			return nil
		}
		return *m
	case *Capture:
		if m == nil {
			return nil
		}
		return *m
	}

	return msg
}

func (c *client) Enqueue(msg Message) (err error) {
	msg = dereferenceMessage(msg)
	if err = msg.Validate(); err != nil {
		return
	}

	var ts = c.now()

	switch m := msg.(type) {
	case Alias:
		m.Type = "alias"
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		m.DisableGeoIP = c.GetDisableGeoIP()
		msg = m

	case Identify:
		m.Type = "identify"
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		m.DisableGeoIP = c.GetDisableGeoIP()
		msg = m

	case GroupIdentify:
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		m.DisableGeoIP = c.GetDisableGeoIP()
		msg = m

	case Capture:
		m.Type = "capture"
		m.Timestamp = makeTimestamp(m.Timestamp, ts)
		if m.SendFeatureFlags {
			// Add all feature variants to event
			featureVariants, err := c.getFeatureVariants(m.DistinctId, m.Groups, NewProperties(), map[string]Properties{})
			if err != nil {
				c.Errorf("unable to get feature variants - %s", err)
			}

			if m.Properties == nil {
				m.Properties = NewProperties()
			}

			for feature, variant := range featureVariants {
				propKey := fmt.Sprintf("$feature/%s", feature)
				m.Properties[propKey] = variant
			}
			// Add all feature flag keys to $active_feature_flags key
			featureKeys := make([]string, len(featureVariants))
			i := 0
			for k := range featureVariants {
				featureKeys[i] = k
				i++
			}
			m.Properties["$active_feature_flags"] = featureKeys
		}
		if m.Properties == nil {
			m.Properties = NewProperties()
		}
		m.Properties.Merge(c.DefaultEventProperties)
		c.setLastCapturedEvent(m)
		msg = m

	default:
		err = fmt.Errorf("messages with custom types cannot be enqueued: %T", msg)
		return
	}

	defer func() {
		// When the `msgs` channel is closed writing to it will trigger a panic.
		// To avoid letting the panic propagate to the caller we recover from it
		// and instead report that the client has been closed and shouldn't be
		// used anymore.
		if recover() != nil {
			err = ErrClosed
		}
	}()

	c.msgs <- msg.APIfy()

	return
}

func (c *client) setLastCapturedEvent(event Capture) {
	c.lastEventMutex.Lock()
	defer c.lastEventMutex.Unlock()
	c.lastCapturedEvent = &event
}

func (c *client) GetLastCapturedEvent() *Capture {
	c.lastEventMutex.RLock()
	defer c.lastEventMutex.RUnlock()
	if c.lastCapturedEvent == nil {
		return nil
	}
	// Return a copy to avoid data races
	eventCopy := *c.lastCapturedEvent
	return &eventCopy
}

func (c *client) IsFeatureEnabled(flagConfig FeatureFlagPayload) (interface{}, error) {
	if err := flagConfig.validate(); err != nil {
		return false, err
	}

	result, err := c.GetFeatureFlag(flagConfig)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *client) ReloadFeatureFlags() error {
	if c.featureFlagsPoller == nil {
		errorMessage := "specifying a PersonalApiKey is required for using feature flags"
		c.Errorf(errorMessage)
		return errors.New(errorMessage)
	}
	c.featureFlagsPoller.ForceReload()
	return nil
}

func (c *client) GetFeatureFlagPayload(flagConfig FeatureFlagPayload) (string, error) {
	if err := flagConfig.validate(); err != nil {
		return "", err
	}

	var payload string
	var err error

	if c.featureFlagsPoller != nil {
		// get feature flag from the poller, which uses the personal api key
		// this is only available when using a PersonalApiKey
		payload, err = c.featureFlagsPoller.GetFeatureFlagPayload(flagConfig)
	} else {
		// if there's no poller, get the feature flag from the flags endpoint
		c.debugf("getting feature flag from flags endpoint")
		payload, err = c.getFeatureFlagPayloadFromRemote(flagConfig.Key, flagConfig.DistinctId, flagConfig.Groups, flagConfig.PersonProperties, flagConfig.GroupProperties)
	}

	return payload, err
}

func (c *client) GetFeatureFlag(flagConfig FeatureFlagPayload) (interface{}, error) {
	if err := flagConfig.validate(); err != nil {
		return false, err
	}

	var flagValue interface{}
	var err error
	var requestId *string
	var flagDetail *FlagDetail

	if c.featureFlagsPoller != nil {
		// get feature flag from the poller, which uses the personal api key
		// this is only available when using a PersonalApiKey
		flagValue, err = c.featureFlagsPoller.GetFeatureFlag(flagConfig)
	} else {
		// if there's no poller, get the feature flag from the flags endpoint
		c.debugf("getting feature flag from flags endpoint")
		flagValue, requestId, err = c.getFeatureFlagFromRemote(flagConfig.Key, flagConfig.DistinctId, flagConfig.Groups,
			flagConfig.PersonProperties, flagConfig.GroupProperties)
		if f, ok := flagValue.(FlagDetail); ok {
			flagValue = f.GetValue()
			flagDetail = &f
		}
	}

	cacheKey := flagUser{flagConfig.DistinctId, flagConfig.Key}
	if *flagConfig.SendFeatureFlagEvents && !c.distinctIdsFeatureFlagsReported.Contains(cacheKey) {
		var properties = NewProperties().
			Set("$feature_flag", flagConfig.Key).
			Set("$feature_flag_response", flagValue).
			Set("$feature_flag_errored", err != nil)

		if requestId != nil {
			properties.Set("$feature_flag_request_id", *requestId)
		}

		if flagDetail != nil {
			properties.Set("$feature_flag_version", flagDetail.Metadata.Version)
			properties.Set("$feature_flag_id", flagDetail.Metadata.ID)
			if flagDetail.Reason != nil {
				properties.Set("$feature_flag_reason", flagDetail.Reason.Description)
			}
		}

		if c.Enqueue(Capture{
			DistinctId: flagConfig.DistinctId,
			Event:      "$feature_flag_called",
			Properties: properties,
			Groups:     flagConfig.Groups,
		}) == nil {
			c.distinctIdsFeatureFlagsReported.Add(cacheKey, struct{}{})
		}
	}

	return flagValue, err
}

func (c *client) GetRemoteConfigPayload(flagKey string) (string, error) {
	return c.makeRemoteConfigRequest(flagKey)
}

// GetFeatureFlags returns all feature flag definitions used for local evaluation
// This is only available when using a PersonalApiKey. Not to be confused with
// GetAllFlags, which returns all flags and their values for a given user.
func (c *client) GetFeatureFlags() ([]FeatureFlag, error) {
	if c.featureFlagsPoller == nil {
		errorMessage := "specifying a PersonalApiKey is required for using feature flags"
		c.Errorf(errorMessage)
		return nil, errors.New(errorMessage)
	}
	return c.featureFlagsPoller.GetFeatureFlags()
}

// GetAllFlags returns all flags and their values for a given user
// A flag value is either a boolean or a variant string (for multivariate flags)
// This first attempts local evaluation if a poller exists, otherwise it falls
// back to the flags endpoint
func (c *client) GetAllFlags(flagConfig FeatureFlagPayloadNoKey) (map[string]interface{}, error) {
	if err := flagConfig.validate(); err != nil {
		return nil, err
	}

	var flagsValue map[string]interface{}
	var err error

	if c.featureFlagsPoller != nil {
		// get feature flags from the poller, which uses the personal api key
		// this is only available when using a PersonalApiKey
		flagsValue, err = c.featureFlagsPoller.GetAllFlags(flagConfig)
	} else {
		// if there's no poller, get the feature flags from the flags endpoint
		c.debugf("getting all feature flags from flags endpoint")
		flagsValue, err = c.getAllFeatureFlagsFromRemote(flagConfig.DistinctId, flagConfig.Groups,
			flagConfig.PersonProperties, flagConfig.GroupProperties)
	}

	return flagsValue, err
}

// Close and flush metrics.
func (c *client) Close() (err error) {
	defer func() {
		// Always recover, a panic could be raised if `c`.quit was closed which
		// means the method was called more than once.
		if recover() != nil {
			err = ErrClosed
		}
	}()
	close(c.quit)
	<-c.shutdown
	return
}

// Asynchronously send a batched requests.
func (c *client) sendAsync(msgs []message, wg *sync.WaitGroup, ex *executor) {
	wg.Add(1)

	if !ex.do(func() {
		defer wg.Done()
		defer func() {
			// In case a bug is introduced in the send function that triggers
			// a panic, we don't want this to ever crash the application so we
			// catch it here and log it instead.
			if err := recover(); err != nil {
				c.Errorf("panic - %s", err)
			}
		}()
		c.send(msgs)
	}) {
		wg.Done()
		c.Errorf("sending messages failed - %s", ErrTooManyRequests)
		c.notifyFailure(msgs, ErrTooManyRequests)
	}
}

// Send batch request.
func (c *client) send(msgs []message) {
	const attempts = 10

	b, err := json.Marshal(batch{
		ApiKey:              c.key,
		HistoricalMigration: c.HistoricalMigration,
		Messages:            msgs,
	})

	if err != nil {
		c.Errorf("marshalling messages - %s", err)
		c.notifyFailure(msgs, err)
		return
	}

	for i := 0; i != attempts; i++ {
		if err = c.upload(b); err == nil {
			c.notifySuccess(msgs)
			return
		}

		// Wait for either a retry timeout or the client to be closed.
		select {
		case <-time.After(c.RetryAfter(i)):
		case <-c.quit:
			c.Errorf("%d messages dropped because they failed to be sent and the client was closed", len(msgs))
			c.notifyFailure(msgs, err)
			return
		}
	}

	c.Errorf("%d messages dropped because they failed to be sent after %d attempts", len(msgs), attempts)
	c.notifyFailure(msgs, err)
}

// Upload serialized batch message.
func (c *client) upload(b []byte) error {
	url := c.Endpoint + "/batch/"
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		c.Errorf("creating request - %s", err)
		return err
	}

	version := getVersion()

	req.Header.Add("User-Agent", SDKName+"/"+version)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", fmt.Sprintf("%d", len(b)))

	res, err := c.http.Do(req)

	if err != nil {
		c.Errorf("sending request - %s", err)
		return err
	}

	defer res.Body.Close()
	return c.report(res)
}

// Report on response body.
func (c *client) report(res *http.Response) (err error) {
	var body []byte

	if res.StatusCode < 300 {
		c.debugf("response %s", res.Status)
		return
	}

	if body, err = ioutil.ReadAll(res.Body); err != nil {
		c.Errorf("response %d %s - %s", res.StatusCode, res.Status, err)
		return
	}

	c.logf("response %d %s – %s", res.StatusCode, res.Status, string(body))
	return fmt.Errorf("%d %s", res.StatusCode, res.Status)
}

// Batch loop.
func (c *client) loop() {
	defer close(c.shutdown)
	if c.featureFlagsPoller != nil {
		defer c.featureFlagsPoller.shutdownPoller()
	}

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	tick := time.NewTicker(c.Interval)
	defer tick.Stop()

	ex := newExecutor(c.maxConcurrentRequests)
	defer ex.close()

	mq := messageQueue{
		maxBatchSize:  c.BatchSize,
		maxBatchBytes: c.maxBatchBytes(),
	}

	for {
		select {
		case msg := <-c.msgs:
			c.push(&mq, msg, wg, ex)

		case <-tick.C:
			c.flush(&mq, wg, ex)

		case <-c.quit:
			c.debugf("exit requested – draining messages")

			// Drain the msg channel, we have to close it first so no more
			// messages can be pushed and otherwise the loop would never end.
			close(c.msgs)
			for msg := range c.msgs {
				c.push(&mq, msg, wg, ex)
			}

			c.flush(&mq, wg, ex)
			c.debugf("exit")
			return
		}
	}
}

func (c *client) push(q *messageQueue, m APIMessage, wg *sync.WaitGroup, ex *executor) {
	var msg message
	var err error

	if msg, err = makeMessage(m, maxMessageBytes); err != nil {
		c.Errorf("%s - %v", err, m)
		c.notifyFailure([]message{{m, nil}}, err)
		return
	}

	c.debugf("buffer (%d/%d) %v", len(q.pending), c.BatchSize, m)

	if msgs := q.push(msg); msgs != nil {
		c.debugf("exceeded messages batch limit with batch of %d messages – flushing", len(msgs))
		c.sendAsync(msgs, wg, ex)
	}
}

func (c *client) flush(q *messageQueue, wg *sync.WaitGroup, ex *executor) {
	if msgs := q.flush(); msgs != nil {
		c.debugf("flushing %d messages", len(msgs))
		c.sendAsync(msgs, wg, ex)
	}
}

func (c *client) debugf(format string, args ...interface{}) {
	if c.Verbose {
		c.logf(format, args...)
	}
}

func (c *client) logf(format string, args ...interface{}) {
	c.Logger.Logf(format, args...)
}

func (c *client) Errorf(format string, args ...interface{}) {
	c.Logger.Errorf(format, args...)
}

func (c *client) maxBatchBytes() int {
	b, _ := json.Marshal(batch{
		Messages: []message{},
	})
	return maxBatchBytes - len(b)
}

func (c *client) notifySuccess(msgs []message) {
	if c.Callback != nil {
		for _, m := range msgs {
			c.Callback.Success(m.msg)
		}
	}
}

func (c *client) notifyFailure(msgs []message, err error) {
	if c.Callback != nil {
		for _, m := range msgs {
			c.Callback.Failure(m.msg, err)
		}
	}
}

func (c *client) getFeatureVariants(distinctId string, groups Groups, personProperties Properties, groupProperties map[string]Properties) (map[string]interface{}, error) {
	if c.featureFlagsPoller == nil {
		errorMessage := "specifying a PersonalApiKey is required for using feature flags"
		c.Errorf(errorMessage)
		return nil, errors.New(errorMessage)
	}

	featureVariants, err := c.featureFlagsPoller.getFeatureFlagVariants(distinctId, groups, personProperties, groupProperties)
	if err != nil {
		return nil, err
	}
	return featureVariants.FeatureFlags, nil
}

func (c *client) makeRemoteConfigRequest(flagKey string) (string, error) {
	remoteConfigEndpoint := fmt.Sprintf("api/projects/@current/feature_flags/%s/remote_config/", flagKey)
	url, err := url.Parse(c.Endpoint + "/" + remoteConfigEndpoint)
	if err != nil {
		return "", fmt.Errorf("creating url: %v", err)
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.PersonalApiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "posthog-go/"+Version)

	res, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code from %s: %d", remoteConfigEndpoint, res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response from /remote_config/: %v", err)
	}

	var responseData string
	if err := json.Unmarshal(resBody, &responseData); err != nil {
		return "", fmt.Errorf("error parsing JSON response from /remote_config/: %v", err)
	}
	return responseData, nil
}

// isFeatureFlagsQuotaLimited checks if feature flags are quota limited in the flags response
func (c *client) isFeatureFlagsQuotaLimited(flagsResponse *FlagsResponse) bool {
	if flagsResponse.QuotaLimited != nil {
		for _, limitedFeature := range *flagsResponse.QuotaLimited {
			if limitedFeature == "feature_flags" {
				c.Errorf("[FEATURE FLAGS] PostHog feature flags quota limited. Learn more about billing limits at https://posthog.com/docs/billing/limits-alerts")
				return true
			}
		}
	}
	return false
}

func (c *client) getFeatureFlagFromRemote(key string, distinctId string, groups Groups, personProperties Properties,
	groupProperties map[string]Properties) (interface{}, *string, error) {

	flagsResponse, err := c.decider.makeFlagsRequest(distinctId, groups, personProperties, groupProperties, c.GetDisableGeoIP())

	if err != nil {
		return nil, nil, err
	}

	if flagsResponse == nil {
		return nil, nil, nil // Should never happen, but just in case. Also helps with type inference.
	}
	var requestId = &flagsResponse.RequestId

	if c.isFeatureFlagsQuotaLimited(flagsResponse) {
		return false, requestId, nil
	}

	if flagDetail, ok := flagsResponse.Flags[key]; ok {
		return flagDetail, requestId, nil
	}

	return false, requestId, nil
}

func (c *client) getFeatureFlagPayloadFromRemote(key string, distinctId string, groups Groups, personProperties Properties, groupProperties map[string]Properties) (string, error) {
	flagsResponse, err := c.decider.makeFlagsRequest(distinctId, groups, personProperties, groupProperties, c.GetDisableGeoIP())
	if err != nil {
		return "", err
	}

	if c.isFeatureFlagsQuotaLimited(flagsResponse) {
		return "", nil
	}

	if value, ok := flagsResponse.FeatureFlagPayloads[key]; ok {
		return value, nil
	}

	return "", nil
}

func (c *client) getAllFeatureFlagsFromRemote(distinctId string, groups Groups, personProperties Properties, groupProperties map[string]Properties) (map[string]interface{}, error) {
	flagsResponse, err := c.decider.makeFlagsRequest(distinctId, groups, personProperties, groupProperties, c.GetDisableGeoIP())
	if err != nil {
		return nil, err
	}

	if c.isFeatureFlagsQuotaLimited(flagsResponse) {
		return map[string]interface{}{}, nil
	}

	return flagsResponse.FeatureFlags, nil
}
