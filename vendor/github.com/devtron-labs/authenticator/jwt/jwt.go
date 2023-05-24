/*
 * Copyright (c) 2021 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Some of the code has been taken from argocd, for them argocd licensing terms apply
 */

package jwt

import (
	"encoding/json"
	"fmt"

	jwt "github.com/golang-jwt/jwt/v4"
)

// MapClaims converts a jwt.Claims to a MapClaims
func MapClaims(claims jwt.Claims) (jwt.MapClaims, error) {
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}
	var mapClaims jwt.MapClaims
	err = json.Unmarshal(claimsBytes, &mapClaims)
	if err != nil {
		return nil, err
	}
	return mapClaims, nil
}

// GetField extracts a field from the claims as a string
func GetField(claims jwt.MapClaims, fieldName string) string {
	if fieldIf, ok := claims[fieldName]; ok {
		if field, ok := fieldIf.(string); ok {
			return field
		}
	}
	return ""
}

// GetScopeValues extracts the values of specified scopes from the claims
func GetScopeValues(claims jwt.MapClaims, scopes []string) []string {
	groups := make([]string, 0)
	for i := range scopes {
		scopeIf, ok := claims[scopes[i]]
		if !ok {
			continue
		}

		switch val := scopeIf.(type) {
		case []interface{}:
			for _, groupIf := range val {
				group, ok := groupIf.(string)
				if ok {
					groups = append(groups, group)
				}
			}
		case []string:
			groups = append(groups, val...)
		case string:
			groups = append(groups, val)
		}
	}

	return groups
}

// GetIssuedAt returns the issued at as an int64
func GetIssuedAt(m jwt.MapClaims) (int64, error) {
	switch iat := m["iat"].(type) {
	case float64:
		return int64(iat), nil
	case json.Number:
		return iat.Int64()
	case int64:
		return iat, nil
	default:
		return 0, fmt.Errorf("iat '%v' is not a number", iat)
	}
}

func Claims(in interface{}) jwt.Claims {
	claims, ok := in.(jwt.Claims)
	if ok {
		return claims
	}
	return nil
}

// IsMember returns whether or not the user's claims is a member of any of the groups
func IsMember(claims jwt.Claims, groups []string) bool {
	mapClaims, err := MapClaims(claims)
	if err != nil {
		return false
	}
	// TODO: groups is hard-wired but we should really be honoring the 'scopes' section in argocd-rbac-cm.
	// O(n^2) loop
	for _, userGroup := range GetScopeValues(mapClaims, []string{"groups"}) {
		for _, group := range groups {
			if userGroup == group {
				return true
			}
		}
	}
	return false
}
