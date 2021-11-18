package util

import "testing"

func TestCompareLimit(t *testing.T) {
	var dat map[string]interface{}
	var err error

	//first test case
	dat = nil
	_, err = CompareLimitsRequests(dat)
	if err != nil {
		t.Errorf("CompareLimitsRequests(nil) gives err = %s, want = nil", err)
	}

	//second test case
	var limits interface{}
	dat = map[string]interface{}{"resources": limits}
	_, err = CompareLimitsRequests(dat)
	if err != nil {
		t.Errorf("CompareLimitsRequests(nil) gives err = %s, want = nil", err)
	}

	//third test case
	var resources interface{}
	resources = map[string]interface{}{"resources": limits}
	dat = map[string]interface{}{"envoyproxy": resources}
	_, err = CompareLimitsRequests(dat)
	if err != nil {
		t.Errorf("CompareLimitsRequests(nil) gives err = %s, want = nil", err)
	}
}
