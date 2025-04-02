package posthog

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const LONG_SCALE = 0xfffffffffffffff

type FeatureFlagsPoller struct {
	loaded         chan bool
	shutdown       chan bool
	forceReload    chan bool
	featureFlags   []FeatureFlag
	cohorts        map[string]PropertyGroup
	groups         map[string]string
	personalApiKey string
	projectApiKey  string
	Errorf         func(format string, args ...interface{})
	Endpoint       string
	http           http.Client
	mutex          sync.RWMutex
	nextPollTick   func() time.Duration
	flagTimeout    time.Duration
}

type FeatureFlag struct {
	Key                        string `json:"key"`
	IsSimpleFlag               bool   `json:"is_simple_flag"`
	RolloutPercentage          *uint8 `json:"rollout_percentage"`
	Active                     bool   `json:"active"`
	Filters                    Filter `json:"filters"`
	EnsureExperienceContinuity *bool  `json:"ensure_experience_continuity"`
}

type Filter struct {
	AggregationGroupTypeIndex *uint8                 `json:"aggregation_group_type_index"`
	Groups                    []FeatureFlagCondition `json:"groups"`
	Multivariate              *Variants              `json:"multivariate"`
	Payloads                  map[string]string      `json:"payloads"`
}

type Variants struct {
	Variants []FlagVariant `json:"variants"`
}

type FlagVariant struct {
	Key               string `json:"key"`
	Name              string `json:"name"`
	RolloutPercentage *uint8 `json:"rollout_percentage"`
}

type FeatureFlagCondition struct {
	Properties        []FlagProperty `json:"properties"`
	RolloutPercentage *uint8         `json:"rollout_percentage"`
	Variant           *string        `json:"variant"`
}

type FlagProperty struct {
	Key      string      `json:"key"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Type     string      `json:"type"`
	Negation bool        `json:"negation"`
}

type PropertyGroup struct {
	Type string `json:"type"`
	// []PropertyGroup or []FlagProperty
	Values []any `json:"values"`
}

type FlagVariantMeta struct {
	ValueMin float64
	ValueMax float64
	Key      string
}

type FeatureFlagsResponse struct {
	Flags            []FeatureFlag            `json:"flags"`
	GroupTypeMapping *map[string]string       `json:"group_type_mapping"`
	Cohorts          map[string]PropertyGroup `json:"cohorts"`
}

type DecideRequestData struct {
	ApiKey           string                `json:"api_key"`
	DistinctId       string                `json:"distinct_id"`
	Groups           Groups                `json:"groups"`
	PersonProperties Properties            `json:"person_properties"`
	GroupProperties  map[string]Properties `json:"group_properties"`
}

type DecideResponse struct {
	FeatureFlags        map[string]interface{} `json:"featureFlags"`
	FeatureFlagPayloads map[string]string      `json:"featureFlagPayloads"`
	QuotaLimited        *[]string              `json:"quota_limited"`
}

type InconclusiveMatchError struct {
	msg string
}

func (e *InconclusiveMatchError) Error() string {
	return e.msg
}

func newFeatureFlagsPoller(
	projectApiKey string,
	personalApiKey string,
	errorf func(format string, args ...interface{}),
	endpoint string,
	httpClient http.Client,
	pollingInterval time.Duration,
	nextPollTick func() time.Duration,
	flagTimeout time.Duration,
) *FeatureFlagsPoller {

	if nextPollTick == nil {
		nextPollTick = func() time.Duration { return pollingInterval }
	}

	poller := FeatureFlagsPoller{
		loaded:         make(chan bool),
		shutdown:       make(chan bool),
		forceReload:    make(chan bool),
		personalApiKey: personalApiKey,
		projectApiKey:  projectApiKey,
		Errorf:         errorf,
		Endpoint:       endpoint,
		http:           httpClient,
		mutex:          sync.RWMutex{},
		nextPollTick:   nextPollTick,
		flagTimeout:    flagTimeout,
	}

	go poller.run()
	return &poller
}

func (poller *FeatureFlagsPoller) run() {
	poller.fetchNewFeatureFlags()
	close(poller.loaded)

	for {
		timer := time.NewTimer(poller.nextPollTick())
		select {
		case <-poller.shutdown:
			close(poller.shutdown)
			close(poller.forceReload)
			timer.Stop()
			return
		case <-poller.forceReload:
			timer.Stop()
			poller.fetchNewFeatureFlags()
		case <-timer.C:
			poller.fetchNewFeatureFlags()
		}
	}
}

func (poller *FeatureFlagsPoller) fetchNewFeatureFlags() {
	personalApiKey := poller.personalApiKey
	headers := [][2]string{{"Authorization", "Bearer " + personalApiKey + ""}}
	res, cancel, err := poller.localEvaluationFlags(headers)
	defer cancel()
	if err != nil {
		poller.Errorf("Unable to fetch feature flags: %s", err)
		return
	}

	// Handle quota limit response (HTTP 402)
	if res.StatusCode == http.StatusPaymentRequired {
		// Clear existing flags when quota limited
		poller.mutex.Lock()
		poller.featureFlags = []FeatureFlag{}
		poller.cohorts = map[string]PropertyGroup{}
		poller.groups = map[string]string{}
		poller.mutex.Unlock()
		poller.Errorf("[FEATURE FLAGS] PostHog feature flags quota limited, resetting feature flag data. Learn more about billing limits at https://posthog.com/docs/billing/limits-alerts")
		return
	}

	if res.StatusCode != http.StatusOK {
		poller.Errorf("Unable to fetch feature flags, status: %s", res.Status)
		return
	}

	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		poller.Errorf("Unable to fetch feature flags: %s", err)
		return
	}
	featureFlagsResponse := FeatureFlagsResponse{}
	err = json.Unmarshal([]byte(resBody), &featureFlagsResponse)
	if err != nil {
		poller.Errorf("Unable to unmarshal response from api/feature_flag/local_evaluation: %s", err)
		return
	}
	newFlags := []FeatureFlag{}
	newFlags = append(newFlags, featureFlagsResponse.Flags...)
	poller.mutex.Lock()
	poller.featureFlags = newFlags
	poller.cohorts = featureFlagsResponse.Cohorts
	if featureFlagsResponse.GroupTypeMapping != nil {
		poller.groups = *featureFlagsResponse.GroupTypeMapping
	}
	poller.mutex.Unlock()
}

func (poller *FeatureFlagsPoller) GetFeatureFlag(flagConfig FeatureFlagPayload) (interface{}, error) {
	flag, err := poller.getFeatureFlag(flagConfig)

	var result interface{}

	if flag.Key != "" {
		result, err = poller.computeFlagLocally(
			flag,
			flagConfig.DistinctId,
			flagConfig.Groups,
			flagConfig.PersonProperties,
			flagConfig.GroupProperties,
			poller.cohorts,
		)
	}

	if err != nil {
		poller.Errorf("Unable to compute flag locally (%s) - %s", flag.Key, err)
	}

	if (err != nil || result == nil) && !flagConfig.OnlyEvaluateLocally {

		result, err = poller.getFeatureFlagVariant(flag, flagConfig.Key, flagConfig.DistinctId, flagConfig.Groups, flagConfig.PersonProperties, flagConfig.GroupProperties)
		if err != nil {
			return nil, err
		}
	}

	return result, err
}

func (poller *FeatureFlagsPoller) GetFeatureFlagPayload(flagConfig FeatureFlagPayload) (string, error) {
	flag, err := poller.getFeatureFlag(flagConfig)

	var variant interface{}

	if flag.Key != "" {
		variant, err = poller.computeFlagLocally(
			flag,
			flagConfig.DistinctId,
			flagConfig.Groups,
			flagConfig.PersonProperties,
			flagConfig.GroupProperties,
			poller.cohorts,
		)
	}
	if err != nil {
		poller.Errorf("Unable to compute flag locally (%s) - %s", flag.Key, err)
	}

	if variant != nil {
		payload, ok := flag.Filters.Payloads[fmt.Sprintf("%v", variant)]
		if ok {
			return payload, nil
		}
	}

	if (variant == nil || err != nil) && !flagConfig.OnlyEvaluateLocally {
		result, err := poller.getFeatureFlagPayload(flagConfig.Key, flagConfig.DistinctId, flagConfig.Groups, flagConfig.PersonProperties, flagConfig.GroupProperties)
		if err != nil {
			return "", err
		}

		return result, nil
	}

	return "", errors.New("unable to compute flag locally")
}

func (poller *FeatureFlagsPoller) getFeatureFlag(flagConfig FeatureFlagPayload) (FeatureFlag, error) {
	featureFlags, err := poller.GetFeatureFlags()
	if err != nil {
		return FeatureFlag{}, err
	}

	featureFlag := FeatureFlag{Key: ""}

	// avoid using flag for conflicts with Golang's stdlib `flag`
	for _, storedFlag := range featureFlags {
		if flagConfig.Key == storedFlag.Key {
			featureFlag = storedFlag
			break
		}
	}

	return featureFlag, nil
}

func (poller *FeatureFlagsPoller) GetAllFlags(flagConfig FeatureFlagPayloadNoKey) (map[string]interface{}, error) {
	response := map[string]interface{}{}
	featureFlags, err := poller.GetFeatureFlags()
	if err != nil {
		return nil, err
	}
	fallbackToDecide := false
	cohorts := poller.cohorts

	if len(featureFlags) == 0 {
		fallbackToDecide = true
	} else {
		for _, storedFlag := range featureFlags {
			result, err := poller.computeFlagLocally(
				storedFlag,
				flagConfig.DistinctId,
				flagConfig.Groups,
				flagConfig.PersonProperties,
				flagConfig.GroupProperties,
				cohorts,
			)
			if err != nil {
				poller.Errorf("Unable to compute flag locally (%s) - %s", storedFlag.Key, err)
				fallbackToDecide = true
			} else {
				response[storedFlag.Key] = result
			}
		}
	}

	if fallbackToDecide && !flagConfig.OnlyEvaluateLocally {
		result, err := poller.getFeatureFlagVariants(flagConfig.DistinctId, flagConfig.Groups, flagConfig.PersonProperties, flagConfig.GroupProperties)

		if err != nil {
			return response, err
		} else {
			for k, v := range result.FeatureFlags {
				response[k] = v
			}
		}
	}

	return response, nil
}

func (poller *FeatureFlagsPoller) computeFlagLocally(
	flag FeatureFlag,
	distinctId string,
	groups Groups,
	personProperties Properties,
	groupProperties map[string]Properties,
	cohorts map[string]PropertyGroup,
) (interface{}, error) {
	if flag.EnsureExperienceContinuity != nil && *flag.EnsureExperienceContinuity {
		return nil, &InconclusiveMatchError{"Flag has experience continuity enabled"}
	}

	if !flag.Active {
		return false, nil
	}

	if flag.Filters.AggregationGroupTypeIndex != nil {

		groupName, exists := poller.groups[fmt.Sprintf("%d", *flag.Filters.AggregationGroupTypeIndex)]

		if !exists {
			errMessage := "flag has unknown group type index"
			return nil, errors.New(errMessage)
		}

		_, exists = groups[groupName]

		if !exists {
			errMessage := fmt.Sprintf("FEATURE FLAGS] Can't compute group feature flag: %s without group names passed in", flag.Key)
			return nil, errors.New(errMessage)
		}

		focusedGroupProperties := groupProperties[groupName]
		return matchFeatureFlagProperties(flag, groups[groupName].(string), focusedGroupProperties, cohorts)
	} else {
		return matchFeatureFlagProperties(flag, distinctId, personProperties, cohorts)
	}
}

func getMatchingVariant(flag FeatureFlag, distinctId string) (interface{}, error) {
	lookupTable := getVariantLookupTable(flag)

	hashValue, err := _hash(flag.Key, distinctId, "variant")
	if err != nil {
		return nil, err
	}

	for _, variant := range lookupTable {
		if hashValue >= float64(variant.ValueMin) && hashValue < float64(variant.ValueMax) {
			return variant.Key, nil
		}
	}

	return true, nil
}

func getVariantLookupTable(flag FeatureFlag) []FlagVariantMeta {
	lookupTable := []FlagVariantMeta{}
	valueMin := 0.00

	multivariates := flag.Filters.Multivariate

	if multivariates == nil || multivariates.Variants == nil {
		return lookupTable
	}

	for _, variant := range multivariates.Variants {
		valueMax := float64(valueMin) + float64(*variant.RolloutPercentage)/100
		_flagVariantMeta := FlagVariantMeta{ValueMin: float64(valueMin), ValueMax: valueMax, Key: variant.Key}
		lookupTable = append(lookupTable, _flagVariantMeta)
		valueMin = float64(valueMax)
	}

	return lookupTable
}

func matchFeatureFlagProperties(
	flag FeatureFlag,
	distinctId string,
	properties Properties,
	cohorts map[string]PropertyGroup,
) (interface{}, error) {
	conditions := flag.Filters.Groups
	isInconclusive := false

	// # Stable sort conditions with variant overrides to the top. This ensures that if overrides are present, they are
	// # evaluated first, and the variant override is applied to the first matching condition.
	// conditionsCopy := make([]PropertyGroup, len(conditions))
	sortedConditions := append([]FeatureFlagCondition{}, conditions...)

	sort.SliceStable(sortedConditions, func(i, j int) bool {
		iValue := 1
		jValue := 1
		if sortedConditions[i].Variant != nil {
			iValue = -1
		}

		if sortedConditions[j].Variant != nil {
			jValue = -1
		}

		return iValue < jValue
	})

	for _, condition := range sortedConditions {

		isMatch, err := isConditionMatch(flag, distinctId, condition, properties, cohorts)
		if err != nil {
			if _, ok := err.(*InconclusiveMatchError); ok {
				isInconclusive = true
			} else {
				return nil, err
			}
		}

		if isMatch {
			variantOverride := condition.Variant
			multivariates := flag.Filters.Multivariate

			if variantOverride != nil && multivariates != nil && multivariates.Variants != nil && containsVariant(multivariates.Variants, *variantOverride) {
				return *variantOverride, nil
			} else {
				return getMatchingVariant(flag, distinctId)
			}
		}
	}

	if isInconclusive {
		return false, &InconclusiveMatchError{"Can't determine if feature flag is enabled or not with given properties"}
	}

	return false, nil
}

func isConditionMatch(
	flag FeatureFlag,
	distinctId string,
	condition FeatureFlagCondition,
	properties Properties,
	cohorts map[string]PropertyGroup,
) (bool, error) {
	if len(condition.Properties) > 0 {
		for _, prop := range condition.Properties {
			var isMatch bool
			var err error

			if prop.Type == "cohort" {
				isMatch, err = matchCohort(prop, properties, cohorts)
			} else {
				isMatch, err = matchProperty(prop, properties)
			}

			if err != nil {
				return false, err
			}

			if !isMatch {
				return false, nil
			}
		}

		if condition.RolloutPercentage != nil {
			return true, nil
		}
	}

	if condition.RolloutPercentage != nil {
		return checkIfSimpleFlagEnabled(flag.Key, distinctId, *condition.RolloutPercentage)
	}

	return true, nil
}

func matchCohort(property FlagProperty, properties Properties, cohorts map[string]PropertyGroup) (bool, error) {
	cohortId := fmt.Sprint(property.Value)
	propertyGroup, ok := cohorts[cohortId]
	if !ok {
		return false, fmt.Errorf("can't match cohort: cohort %s not found", cohortId)
	}

	return matchPropertyGroup(propertyGroup, properties, cohorts)
}

func matchPropertyGroup(propertyGroup PropertyGroup, properties Properties, cohorts map[string]PropertyGroup) (bool, error) {
	groupType := propertyGroup.Type
	values := propertyGroup.Values

	if len(values) == 0 {
		// empty groups are no-ops, always match
		return true, nil
	}

	errorMatchingLocally := false

	for _, value := range values {
		switch prop := value.(type) {
		case map[string]any:
			if _, ok := prop["values"]; ok {
				// PropertyGroup
				matches, err := matchPropertyGroup(PropertyGroup{
					Type:   getSafeProp[string](prop, "type"),
					Values: getSafeProp[[]any](prop, "values"),
				}, properties, cohorts)
				if err != nil {
					if _, ok := err.(*InconclusiveMatchError); ok {
						errorMatchingLocally = true
					} else {
						return false, err
					}
				}

				if groupType == "AND" {
					if !matches {
						return false, nil
					}
				} else {
					// OR group
					if matches {
						return true, nil
					}
				}
			} else {
				// FlagProperty
				var matches bool
				var err error
				flagProperty := FlagProperty{
					Key:      getSafeProp[string](prop, "key"),
					Operator: getSafeProp[string](prop, "operator"),
					Value:    getSafeProp[any](prop, "value"),
					Type:     getSafeProp[string](prop, "type"),
					Negation: getSafeProp[bool](prop, "negation"),
				}
				if prop["type"] == "cohort" {
					matches, err = matchCohort(flagProperty, properties, cohorts)
				} else {
					matches, err = matchProperty(flagProperty, properties)
				}

				if err != nil {
					if _, ok := err.(*InconclusiveMatchError); ok {
						errorMatchingLocally = true
					} else {
						return false, err
					}
				}

				negation := flagProperty.Negation
				if groupType == "AND" {
					// if negated property, do the inverse
					if !matches && !negation {
						return false, nil
					}
					if matches && negation {
						return false, nil
					}
				} else {
					// OR group
					if matches && !negation {
						return true, nil
					}
					if !matches && negation {
						return true, nil
					}
				}
			}
		}
	}

	if errorMatchingLocally {
		return false, &InconclusiveMatchError{msg: "Can't match cohort without a given cohort property value"}
	}

	// if we get here, all matched in AND case, or none matched in OR case
	return groupType == "AND", nil
}

func matchProperty(property FlagProperty, properties Properties) (bool, error) {
	key := property.Key
	operator := property.Operator
	value := property.Value
	if _, ok := properties[key]; !ok {
		return false, &InconclusiveMatchError{"Can't match properties without a given property value"}
	}

	if operator == "is_not_set" {
		return false, &InconclusiveMatchError{"Can't match properties with operator is_not_set"}
	}

	override_value := properties[key]

	if operator == "exact" {
		switch t := value.(type) {
		case []interface{}:
			return contains(t, override_value), nil
		default:
			return value == override_value, nil
		}
	}

	if operator == "is_not" {
		switch t := value.(type) {
		case []interface{}:
			return !contains(t, override_value), nil
		default:
			return value != override_value, nil
		}
	}

	if operator == "is_set" {
		return true, nil
	}

	if operator == "icontains" {
		return strings.Contains(strings.ToLower(fmt.Sprintf("%v", override_value)), strings.ToLower(fmt.Sprintf("%v", value))), nil
	}

	if operator == "not_icontains" {
		return !strings.Contains(strings.ToLower(fmt.Sprintf("%v", override_value)), strings.ToLower(fmt.Sprintf("%v", value))), nil
	}

	if operator == "regex" {

		r, err := regexp.Compile(fmt.Sprintf("%v", value))
		// invalid regex
		if err != nil {
			return false, nil
		}

		match := r.MatchString(fmt.Sprintf("%v", override_value))

		if match {
			return true, nil
		} else {
			return false, nil
		}
	}

	if operator == "not_regex" {
		var r *regexp.Regexp
		var err error

		if valueString, ok := value.(string); ok {
			r, err = regexp.Compile(valueString)
		} else if valueInt, ok := value.(int); ok {
			valueString = strconv.Itoa(valueInt)
			r, err = regexp.Compile(valueString)
		} else {
			errMessage := "regex expression not allowed"
			return false, errors.New(errMessage)
		}

		// invalid regex
		if err != nil {
			return false, nil
		}

		var match bool
		if valueString, ok := override_value.(string); ok {
			match = r.MatchString(valueString)
		} else if valueInt, ok := override_value.(int); ok {
			valueString = strconv.Itoa(valueInt)
			match = r.MatchString(valueString)
		} else {
			errMessage := "value type not supported"
			return false, errors.New(errMessage)
		}

		if !match {
			return true, nil
		} else {
			return false, nil
		}
	}

	if operator == "gt" {
		valueOrderable, overrideValueOrderable, err := validateOrderable(value, override_value)
		if err != nil {
			return false, err
		}

		return overrideValueOrderable > valueOrderable, nil
	}

	if operator == "lt" {
		valueOrderable, overrideValueOrderable, err := validateOrderable(value, override_value)
		if err != nil {
			return false, err
		}

		return overrideValueOrderable < valueOrderable, nil
	}

	if operator == "gte" {
		valueOrderable, overrideValueOrderable, err := validateOrderable(value, override_value)
		if err != nil {
			return false, err
		}

		return overrideValueOrderable >= valueOrderable, nil
	}

	if operator == "lte" {
		valueOrderable, overrideValueOrderable, err := validateOrderable(value, override_value)
		if err != nil {
			return false, err
		}

		return overrideValueOrderable <= valueOrderable, nil
	}

	return false, &InconclusiveMatchError{"Unknown operator: " + operator}

}

func validateOrderable(firstValue interface{}, secondValue interface{}) (float64, float64, error) {
	convertedFirstValue, err := interfaceToFloat(firstValue)
	if err != nil {
		errMessage := "value 1 is not orderable"
		return 0, 0, errors.New(errMessage)
	}
	convertedSecondValue, err := interfaceToFloat(secondValue)
	if err != nil {
		errMessage := "value 2 is not orderable"
		return 0, 0, errors.New(errMessage)
	}

	return convertedFirstValue, convertedSecondValue, nil
}

func interfaceToFloat(val interface{}) (float64, error) {
	var i float64
	switch t := val.(type) {
	case int:
		i = float64(t)
	case int8:
		i = float64(t)
	case int16:
		i = float64(t)
	case int32:
		i = float64(t)
	case int64:
		i = float64(t)
	case float32:
		i = float64(t)
	case float64:
		i = float64(t)
	case uint8:
		i = float64(t)
	case uint16:
		i = float64(t)
	case uint32:
		i = float64(t)
	case uint64:
		i = float64(t)
	default:
		errMessage := "argument not orderable"
		return 0.0, errors.New(errMessage)
	}

	return i, nil
}

func contains(s []interface{}, e interface{}) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func containsVariant(variantList []FlagVariant, key string) bool {
	for _, variant := range variantList {
		if variant.Key == key {
			return true
		}
	}
	return false
}

func (poller *FeatureFlagsPoller) isSimpleFlagEnabled(key string, distinctId string, rolloutPercentage uint8) (bool, error) {
	isEnabled, err := checkIfSimpleFlagEnabled(key, distinctId, rolloutPercentage)
	if err != nil {
		errMessage := "Error converting string to int"
		poller.Errorf(errMessage)
		return false, errors.New(errMessage)
	}
	return isEnabled, nil
}

// extracted as a regular func for testing purposes
func checkIfSimpleFlagEnabled(key string, distinctId string, rolloutPercentage uint8) (bool, error) {
	val, err := _hash(key, distinctId, "")
	if err != nil {
		return false, err
	}

	return val <= float64(rolloutPercentage)/100, nil
}

func _hash(key string, distinctId string, salt string) (float64, error) {
	hash := sha1.New()
	hash.Write([]byte("" + key + "." + distinctId + "" + salt))
	digest := hash.Sum(nil)
	hexString := fmt.Sprintf("%x\n", digest)[:15]

	value, err := strconv.ParseInt(hexString, 16, 64)
	if err != nil {
		return 0, err
	}

	return float64(value) / LONG_SCALE, nil
}

func (poller *FeatureFlagsPoller) GetFeatureFlags() ([]FeatureFlag, error) {
	// When channel is open this will block. When channel is closed it will immediately exit.
	_, closed := <-poller.loaded
	if closed && poller.featureFlags == nil {
		// There was an error with initial flag fetching
		return nil, fmt.Errorf("flags were not successfully fetched yet")
	}

	return poller.featureFlags, nil
}

func (poller *FeatureFlagsPoller) decide(requestData []byte, headers [][2]string) (*http.Response, context.CancelFunc, error) {
	decideEndpoint := "decide/?v=3"

	url, err := url.Parse(poller.Endpoint + "/" + decideEndpoint + "")
	if err != nil {
		poller.Errorf("creating url - %s", err)
	}

	return poller.request("POST", url, requestData, headers, poller.flagTimeout)
}

func (poller *FeatureFlagsPoller) localEvaluationFlags(headers [][2]string) (*http.Response, context.CancelFunc, error) {
	localEvaluationEndpoint := "api/feature_flag/local_evaluation"

	url, err := url.Parse(poller.Endpoint + "/" + localEvaluationEndpoint + "")
	if err != nil {
		poller.Errorf("creating url - %s", err)
	}
	searchParams := url.Query()
	searchParams.Add("token", poller.projectApiKey)
	searchParams.Add("send_cohorts", "true")
	url.RawQuery = searchParams.Encode()

	return poller.request("GET", url, []byte{}, headers, time.Duration(10)*time.Second)
}

func (poller *FeatureFlagsPoller) request(method string, url *url.URL, requestData []byte, headers [][2]string, timeout time.Duration) (*http.Response, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(requestData))
	if err != nil {
		poller.Errorf("creating request - %s", err)
	}

	version := getVersion()

	req.Header.Add("User-Agent", SdkName+"/"+version)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", fmt.Sprintf("%d", len(requestData)))

	for _, header := range headers {
		req.Header.Add(header[0], header[1])
	}

	res, err := poller.http.Do(req)
	if err != nil {
		poller.Errorf("sending request - %s", err)
	}

	return res, cancel, err
}

func (poller *FeatureFlagsPoller) ForceReload() {
	poller.forceReload <- true
}

func (poller *FeatureFlagsPoller) shutdownPoller() {
	poller.shutdown <- true
}

func (poller *FeatureFlagsPoller) getFeatureFlagVariants(distinctId string, groups Groups, personProperties Properties, groupProperties map[string]Properties) (*DecideResponse, error) {
	errorMessage := "Failed when getting flag variants"
	requestDataBytes, err := json.Marshal(DecideRequestData{
		ApiKey:           poller.projectApiKey,
		DistinctId:       distinctId,
		Groups:           groups,
		PersonProperties: personProperties,
		GroupProperties:  groupProperties,
	})
	headers := [][2]string{{"Authorization", "Bearer " + poller.personalApiKey + ""}}
	if err != nil {
		errorMessage = "unable to marshal decide endpoint request data"
		poller.Errorf(errorMessage)
		return nil, errors.New(errorMessage)
	}
	res, cancel, err := poller.decide(requestDataBytes, headers)
	defer cancel()
	if err != nil || res.StatusCode != http.StatusOK {
		errorMessage = "Error calling /decide/"
		if err != nil {
			errorMessage += " - " + err.Error()
		}
		poller.Errorf(errorMessage)
		return nil, errors.New(errorMessage)
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		errorMessage = "Error reading response from /decide/"
		poller.Errorf(errorMessage)
		return nil, errors.New(errorMessage)
	}
	defer res.Body.Close()
	decideResponse := DecideResponse{}
	err = json.Unmarshal([]byte(resBody), &decideResponse)
	if err != nil {
		errorMessage = "Error parsing response from /decide/"
		poller.Errorf(errorMessage)
		return nil, errors.New(errorMessage)
	}

	return &decideResponse, nil
}

func (poller *FeatureFlagsPoller) getFeatureFlagVariant(featureFlag FeatureFlag, key string, distinctId string, groups Groups, personProperties Properties, groupProperties map[string]Properties) (interface{}, error) {
	var result interface{} = false

	if featureFlag.IsSimpleFlag {

		// json.Unmarshal will convert JSON `null` to a nullish value for each type
		// which is 0 for uint. However, our feature flags should have rolloutPercentage == 100
		// if it is set to `null`. Having rollout percentage be a pointer and deferencing it
		// here allows its value to be `nil` following json.Unmarhsal, so we can appropriately
		// set it to 100
		rolloutPercentage := uint8(100)
		if featureFlag.RolloutPercentage != nil {
			rolloutPercentage = *featureFlag.RolloutPercentage
		}
		var err error
		result, err = poller.isSimpleFlagEnabled(key, distinctId, rolloutPercentage)
		if err != nil {
			return false, err
		}
	} else {
		featureFlagVariants, variantErr := poller.getFeatureFlagVariants(distinctId, groups, personProperties, groupProperties)

		if variantErr != nil {
			return false, variantErr
		}

		for flagKey, flagValue := range featureFlagVariants.FeatureFlags {
			flagValueString := fmt.Sprintf("%v", flagValue)
			if key == flagKey && flagValueString != "false" {
				result = flagValueString
				break
			}
		}
		return result, nil
	}
	return result, nil
}

func (poller *FeatureFlagsPoller) getFeatureFlagPayload(key string, distinctId string, groups Groups, personProperties Properties, groupProperties map[string]Properties) (string, error) {
	featureFlagVariants, err := poller.getFeatureFlagVariants(distinctId, groups, personProperties, groupProperties)
	if err != nil {
		return "", err
	}

	return featureFlagVariants.FeatureFlagPayloads[key], nil
}

func getSafeProp[T any](properties map[string]any, key string) T {
	switch v := properties[key].(type) {
	case T:
		return v
	default:
		var defaultValue T
		return defaultValue
	}
}
