/*
 * Copyright (c) 2020 Devtron Labs
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
 */

package security

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

type CvePolicy struct {
	tableName     struct{}     `sql:"cve_policy_control" pg:",discard_unknown_columns"`
	Id            int          `sql:"id,pk"`
	Global        bool         `sql:"global,notnull"`
	ClusterId     int          `sql:"cluster_id"`
	EnvironmentId int          `sql:"env_id"`
	AppId         int          `sql:"app_id"`
	CVEStoreId    string       `sql:"cve_store_id"`
	Action        PolicyAction `sql:"action, notnull"`
	Severity      *Severity    `sql:"severity, notnull "`
	Deleted       bool         `sql:"deleted, notnull"`
	sql.AuditLog
	CveStore *CveStore
}

type PolicyAction int

const (
	Inherit PolicyAction = iota
	Allow
	Block
)

func (d PolicyAction) String() string {
	return [...]string{"inherit", "allow", "block"}[d]
}

//------------------
type Severity int

const (
	Low Severity = iota
	Moderate
	Critical
)

func (d Severity) ValuesOf(severity string) Severity {
	if severity == "critical" {
		return Critical
	} else if severity == "moderate" {
		return Moderate
	} else if severity == "low" {
		return Low
	}
	return Low
}
func (d Severity) String() string {
	return [...]string{"low", "moderate", "critical"}[d]
}

//----------------
type PolicyLevel int

const (
	Global PolicyLevel = iota
	Cluster
	Environment
	Application
)

func (d PolicyLevel) String() string {
	return [...]string{"global", "cluster", "environment", "application"}[d]
}

func (policy *CvePolicy) PolicyLevel() PolicyLevel {
	if policy.ClusterId != 0 {
		return Cluster
	} else if policy.EnvironmentId != 0 {
		return Environment
	} else if policy.AppId != 0 {
		return Application
	} else {
		return Global
	}
}

//------------------

type CvePolicyRepository interface {
	GetGlobalPolicies() (policies []*CvePolicy, err error)
	GetClusterPolicies(clusterId int) (policies []*CvePolicy, err error)
	GetEnvPolicies(clusterId int, environmentId int) (policies []*CvePolicy, err error)
	GetAppEnvPolicies(clusterId int, environmentId int, appId int) (policies []*CvePolicy, err error)
	SavePolicy(policy *CvePolicy) (*CvePolicy, error)
	UpdatePolicy(policy *CvePolicy) (*CvePolicy, error)
	GetById(id int) (*CvePolicy, error)
	GetBlockedCVEList(cves []*CveStore, clusterId, envId, appId int, isAppstore bool) ([]*CveStore, error)
}
type CvePolicyRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewPolicyRepositoryImpl(dbConnection *pg.DB) *CvePolicyRepositoryImpl {
	return &CvePolicyRepositoryImpl{dbConnection: dbConnection}
}
func (impl *CvePolicyRepositoryImpl) GetGlobalPolicies() (policies []*CvePolicy, err error) {
	err = impl.dbConnection.Model(&policies).
		Column("cve_policy.*").
		Relation("CveStore").
		Where("global = true").
		Where("deleted = false").
		Select()
	return policies, err
}

func (impl *CvePolicyRepositoryImpl) GetClusterPolicies(clusterId int) (policies []*CvePolicy, err error) {
	err = impl.dbConnection.Model(&policies).
		Column("cve_policy.*").
		Relation("CveStore").
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q = q.WhereOr("cluster_id = ?", clusterId).
				WhereOr("global = true")
			return q, nil
		}).
		Where("deleted = false").
		Select()
	return policies, err
}

func (impl *CvePolicyRepositoryImpl) GetEnvPolicies(clusterId int, environmentId int) (policies []*CvePolicy, err error) {
	err = impl.dbConnection.Model(&policies).
		Column("cve_policy.*").
		Relation("CveStore").
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q = q.WhereOr("cluster_id = ?", clusterId).
				WhereOr("env_id = ?", environmentId).
				WhereOr("global = true")
			return q, nil
		}).
		Where("deleted = false").
		Select()
	return policies, err
}

func (impl *CvePolicyRepositoryImpl) GetAppEnvPolicies(clusterId int, environmentId int, appId int) (policies []*CvePolicy, err error) {
	err = impl.dbConnection.Model(&policies).
		Column("cve_policy.*").
		Relation("CveStore").
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q = q.WhereOr("cluster_id = ?", clusterId).
				WhereOr("env_id = ?", environmentId).
				WhereOr("global = true").
				WhereOr("app_id = ?", appId)
			return q, nil
		}).
		Where("deleted = false").
		Select()
	return policies, err
}

func (impl *CvePolicyRepositoryImpl) SavePolicy(policy *CvePolicy) (*CvePolicy, error) {
	err := impl.dbConnection.Insert(policy)
	return policy, err
}
func (impl *CvePolicyRepositoryImpl) UpdatePolicy(policy *CvePolicy) (*CvePolicy, error) {
	_, err := impl.dbConnection.Model(policy).WherePK().UpdateNotNull()
	return policy, err
}
func (impl *CvePolicyRepositoryImpl) GetById(id int) (*CvePolicy, error) {
	cvePolicy := &CvePolicy{Id: id}
	err := impl.dbConnection.Model(cvePolicy).WherePK().Select()
	return cvePolicy, err
}

func (impl *CvePolicyRepositoryImpl) GetBlockedCVEList(cves []*CveStore, clusterId, envId, appId int, isAppstore bool) ([]*CveStore, error) {

	cvePolicy, severityPolicy, err := impl.getApplicablePolicy(clusterId, envId, appId, isAppstore)
	if err != nil {
		return nil, err
	}
	blockedCve := impl.enforceCvePolicy(cves, cvePolicy, severityPolicy)
	return blockedCve, nil
}

func (impl *CvePolicyRepositoryImpl) enforceCvePolicy(cves []*CveStore, cvePolicy map[string]*CvePolicy, severityPolicy map[Severity]*CvePolicy) (blockedCVE []*CveStore) {

	for _, cve := range cves {
		if policy, ok := cvePolicy[cve.Name]; ok {
			if policy.Action == Allow {
				continue
			} else {
				blockedCVE = append(blockedCVE, cve)
			}
		} else {
			if severityPolicy[cve.Severity] != nil && severityPolicy[cve.Severity].Action == Allow {
				continue
			} else {
				blockedCVE = append(blockedCVE, cve)
			}
		}
	}
	return blockedCVE
}

func (impl *CvePolicyRepositoryImpl) getApplicablePolicy(clusterId, envId, appId int, isAppstore bool) (map[string]*CvePolicy, map[Severity]*CvePolicy, error) {

	var policyLevel PolicyLevel
	if isAppstore && appId > 0 && envId > 0 && clusterId > 0 {
		policyLevel = Environment
	} else if appId > 0 && envId > 0 && clusterId > 0 {
		policyLevel = Application
	} else if envId > 0 && clusterId > 0 {
		policyLevel = Environment
	} else if clusterId > 0 {
		policyLevel = Cluster
	} else {
		//error in case of global or other policy
		return nil, nil, fmt.Errorf("policy not identified")
	}

	cvePolicy, severityPolicy, err := impl.getPolicies(policyLevel, clusterId, envId, appId)
	return cvePolicy, severityPolicy, err
}

func (impl *CvePolicyRepositoryImpl) getPolicies(policyLevel PolicyLevel, clusterId, environmentId, appId int) (map[string]*CvePolicy, map[Severity]*CvePolicy, error) {
	var policies []*CvePolicy
	var err error
	if policyLevel == Global {
		policies, err = impl.GetGlobalPolicies()
	} else if policyLevel == Cluster {
		policies, err = impl.GetClusterPolicies(clusterId)
	} else if policyLevel == Environment {
		policies, err = impl.GetEnvPolicies(clusterId, environmentId)
	} else if policyLevel == Application {
		policies, err = impl.GetAppEnvPolicies(clusterId, environmentId, appId)
	} else {
		return nil, nil, fmt.Errorf("unsupported policy level: %s", policyLevel)
	}
	if err != nil {
		//impl.logger.Errorw("error in fetching policy  ", "level", policyLevel, "err", err)
		return nil, nil, err
	}
	cvePolicy, severityPolicy := impl.getApplicablePolicies(policies)
	//impl.logger.Debugw("policy identified ", "policyLevel", policyLevel)
	//transform and return
	return cvePolicy, severityPolicy, nil
}

func (impl *CvePolicyRepositoryImpl) getApplicablePolicies(policies []*CvePolicy) (map[string]*CvePolicy, map[Severity]*CvePolicy) {
	cvePolicy := make(map[string][]*CvePolicy)
	severityPolicy := make(map[Severity][]*CvePolicy)
	for _, policy := range policies {
		if policy.CVEStoreId != "" {
			cvePolicy[policy.CveStore.Name] = append(cvePolicy[policy.CveStore.Name], policy)
		} else {
			severityPolicy[*policy.Severity] = append(severityPolicy[*policy.Severity], policy)
		}
	}
	applicableCvePolicy := impl.getHighestPolicy(cvePolicy)
	applicableSeverityPolicy := impl.getHighestPolicyS(severityPolicy)
	return applicableCvePolicy, applicableSeverityPolicy
}

func (impl *CvePolicyRepositoryImpl) getHighestPolicy(allPolicies map[string][]*CvePolicy) map[string]*CvePolicy {
	applicablePolicies := make(map[string]*CvePolicy)
	for key, policies := range allPolicies {
		var applicablePolicy *CvePolicy
		for _, policy := range policies {
			if applicablePolicy == nil {
				applicablePolicy = policy
			} else {
				if policy.PolicyLevel() > applicablePolicy.PolicyLevel() {
					applicablePolicy = policy
				}
			}
		}
		applicablePolicies[key] = applicablePolicy
	}
	return applicablePolicies
}
func (impl *CvePolicyRepositoryImpl) getHighestPolicyS(allPolicies map[Severity][]*CvePolicy) map[Severity]*CvePolicy {
	applicablePolicies := make(map[Severity]*CvePolicy)
	for key, policies := range allPolicies {
		var applicablePolicy *CvePolicy
		for _, policy := range policies {
			if applicablePolicy == nil {
				applicablePolicy = policy
			} else {
				if policy.PolicyLevel() > applicablePolicy.PolicyLevel() {
					applicablePolicy = policy
				}
			}
		}
		applicablePolicies[key] = applicablePolicy
	}
	return applicablePolicies
}
