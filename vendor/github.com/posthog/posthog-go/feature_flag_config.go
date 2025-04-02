package posthog

type FeatureFlagPayload struct {
	Key                   string
	DistinctId            string
	Groups                Groups
	PersonProperties      Properties
	GroupProperties       map[string]Properties
	OnlyEvaluateLocally   bool
	SendFeatureFlagEvents *bool
}

func (c *FeatureFlagPayload) validate() error {
	if len(c.Key) == 0 {
		return ConfigError{
			Reason: "Feature Flag Key required",
			Field:  "Key",
			Value:  c.Key,
		}
	}

	if len(c.DistinctId) == 0 {
		return ConfigError{
			Reason: "DistinctId required",
			Field:  "Distinct Id",
			Value:  c.DistinctId,
		}
	}

	if c.Groups == nil {
		c.Groups = Groups{}
	}

	if c.PersonProperties == nil {
		c.PersonProperties = NewProperties()
	}

	if c.GroupProperties == nil {
		c.GroupProperties = map[string]Properties{}
	}

	if c.SendFeatureFlagEvents == nil {
		tempTrue := true
		c.SendFeatureFlagEvents = &tempTrue
	}
	return nil
}

type FeatureFlagPayloadNoKey struct {
	DistinctId            string
	Groups                Groups
	PersonProperties      Properties
	GroupProperties       map[string]Properties
	OnlyEvaluateLocally   bool
	SendFeatureFlagEvents *bool
}

func (c *FeatureFlagPayloadNoKey) validate() error {
	if len(c.DistinctId) == 0 {
		return ConfigError{
			Reason: "DistinctId required",
			Field:  "Distinct Id",
			Value:  c.DistinctId,
		}
	}

	if c.Groups == nil {
		c.Groups = Groups{}
	}

	if c.PersonProperties == nil {
		c.PersonProperties = NewProperties()
	}

	if c.GroupProperties == nil {
		c.GroupProperties = map[string]Properties{}
	}

	if c.SendFeatureFlagEvents == nil {
		tempTrue := true
		c.SendFeatureFlagEvents = &tempTrue
	}
	return nil
}
