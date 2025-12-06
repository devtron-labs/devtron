/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package repository

import (
	"fmt"
	securityBean "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type CvePolicy struct {
	tableName     struct{}                  `sql:"cve_policy_control" pg:",discard_unknown_columns"`
	Id            int                       `sql:"id,pk"`
	Global        bool                      `sql:"global,notnull,type:boolean"`
	ClusterId     int                       `sql:"cluster_id,type:integer"`
	EnvironmentId int                       `sql:"env_id,type:integer"`
	AppId         int                       `sql:"app_id,type:integer"`
	CVEStoreId    string                    `sql:"cve_store_id,type:text"`
	Action        securityBean.PolicyAction `sql:"action, notnull,type:integer"`
	Severity      *securityBean.Severity    `sql:"severity, notnull,type:integer"`
	Deleted       bool                      `sql:"deleted, notnull,type:boolean"`
	sql.AuditLog
	CveStore *CveStore
}

func (policy *CvePolicy) PolicyLevel() securityBean.PolicyLevel {
	if policy.ClusterId != 0 {
		return securityBean.Cluster
	} else if policy.AppId != 0 {
		return securityBean.Application
	} else if policy.EnvironmentId != 0 {
		return securityBean.Environment
	} else {
		return securityBean.Global
	}
}

func (policy *CvePolicy) UpdateDeleted(deleted bool) {
	policy.Deleted = deleted
}

//------------------

type CvePolicyRepository interface {
	GetGlobalPolicies() (policies []*CvePolicy, err error)
	GetClusterPolicies(clusterId int) (policies []*CvePolicy, err error)
	GetEnvPolicies(clusterId int, environmentId int) (policies []*CvePolicy, err error)
	GetAppEnvPolicies(clusterId int, environmentId int, appId int) (policies []*CvePolicy, err error)
	SavePolicy(tx *pg.Tx, policy *CvePolicy) (*CvePolicy, error)
	UpdatePolicy(tx *pg.Tx, policy *CvePolicy) (*CvePolicy, error)
	GetById(id int) (*CvePolicy, error)
	GetBlockedCVEList(cves []*CveStore, clusterId, envId, appId int, isAppstore bool) ([]*CveStore, error)
	GetActiveByCveIdAndScope(cveId string, envId, appId, clusterId int) (policies []*CvePolicy, err error)
	UpdatePoliciesInBulk(tx *pg.Tx, policies []*CvePolicy) error
}

type CvePolicyRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPolicyRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CvePolicyRepositoryImpl {
	return &CvePolicyRepositoryImpl{dbConnection: dbConnection, logger: logger}
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
		Where("app_id is null").
		Select()
	return policies, err
}

func (impl *CvePolicyRepositoryImpl) GetAppEnvPolicies(clusterId int, environmentId int, appId int) (policies []*CvePolicy, err error) {
	err = impl.dbConnection.Model(&policies).
		Column("cve_policy.*").
		Relation("CveStore").
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q = q.WhereOr("cluster_id = ?", clusterId).
				WhereOrGroup(func(sq *orm.Query) (*orm.Query, error) {
					sq = sq.Where("env_id = ?", environmentId).Where("app_id is null")
					return sq, nil
				}).
				WhereOr("global = true").
				WhereOrGroup(func(sq *orm.Query) (*orm.Query, error) {
					sq = sq.Where("app_id = ?", appId).Where("env_id = ?", environmentId)
					return sq, nil
				}).
				WhereOrGroup(func(sq *orm.Query) (*orm.Query, error) {
					sq = sq.Where("app_id = ?", appId).Where("env_id is null")
					return sq, nil
				})
			return q, nil
		}).
		Where("deleted = false").
		Select()
	return policies, err
}

func (impl *CvePolicyRepositoryImpl) SavePolicy(tx *pg.Tx, policy *CvePolicy) (*CvePolicy, error) {
	var err error
	if tx == nil {
		err = impl.dbConnection.Insert(policy)
		if err != nil {
			impl.logger.Errorw("error in inserting policy", "policy", policy, "err", err)
			return nil, err
		}
	} else {
		err = tx.Insert(policy)
		if err != nil {
			impl.logger.Errorw("error in inserting policy in transaction", "policy", policy, "err", err)
			return nil, err
		}
	}
	return policy, nil
}

func (impl *CvePolicyRepositoryImpl) UpdatePolicy(tx *pg.Tx, policy *CvePolicy) (*CvePolicy, error) {
	var err error
	if tx != nil {
		err = tx.Update(policy)
		if err != nil {
			impl.logger.Errorw("error in updating cvePolicy in transaction", "policy", policy, "err", err)
			return nil, err
		}
	} else {
		_, err = impl.dbConnection.Model(policy).WherePK().UpdateNotNull()
		if err != nil {
			impl.logger.Errorw("error in updating cvePolicy", "policy", policy, "err", err)
			return nil, err
		}
	}
	return policy, nil
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
	blockedCve := EnforceCvePolicy(cves, cvePolicy, severityPolicy)
	return blockedCve, nil
}

func EnforceCvePolicy(cves []*CveStore, cvePolicy map[string]*CvePolicy, severityPolicy map[securityBean.Severity]*CvePolicy) (blockedCVE []*CveStore) {

	for _, cve := range cves {
		if policy, ok := cvePolicy[cve.Name]; ok {
			if policy.Action == securityBean.Allow {
				continue
			} else if (policy.Action == securityBean.Block) || (policy.Action == securityBean.Blockiffixed && cve.FixedVersion != "") {
				blockedCVE = append(blockedCVE, cve)
			}
		} else {
			if severityPolicy[cve.GetSeverity()] != nil && severityPolicy[cve.GetSeverity()].Action == securityBean.Allow {
				continue
			} else if severityPolicy[cve.GetSeverity()] != nil && (severityPolicy[cve.GetSeverity()].Action == securityBean.Block || (severityPolicy[cve.GetSeverity()].Action == securityBean.Blockiffixed && cve.FixedVersion != "")) {
				blockedCVE = append(blockedCVE, cve)
			}
		}
	}
	return blockedCVE
}

func (impl *CvePolicyRepositoryImpl) getApplicablePolicy(clusterId, envId, appId int, isAppstore bool) (map[string]*CvePolicy, map[securityBean.Severity]*CvePolicy, error) {

	var policyLevel securityBean.PolicyLevel
	if isAppstore && appId > 0 && envId > 0 && clusterId > 0 {
		policyLevel = securityBean.Environment
	} else if appId > 0 && envId > 0 && clusterId > 0 {
		policyLevel = securityBean.Application
	} else if envId > 0 && clusterId > 0 {
		policyLevel = securityBean.Environment
	} else if clusterId > 0 {
		policyLevel = securityBean.Cluster
	} else {
		//error in case of global or other policy
		return nil, nil, fmt.Errorf("policy not identified")
	}

	cvePolicy, severityPolicy, err := impl.getPolicies(policyLevel, clusterId, envId, appId)
	return cvePolicy, severityPolicy, err
}

func (impl *CvePolicyRepositoryImpl) getPolicies(policyLevel securityBean.PolicyLevel, clusterId, environmentId, appId int) (map[string]*CvePolicy, map[securityBean.Severity]*CvePolicy, error) {
	var policies []*CvePolicy
	var err error
	if policyLevel == securityBean.Global {
		policies, err = impl.GetGlobalPolicies()
	} else if policyLevel == securityBean.Cluster {
		policies, err = impl.GetClusterPolicies(clusterId)
	} else if policyLevel == securityBean.Environment {
		policies, err = impl.GetEnvPolicies(clusterId, environmentId)
	} else if policyLevel == securityBean.Application {
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

func (impl *CvePolicyRepositoryImpl) getApplicablePolicies(policies []*CvePolicy) (map[string]*CvePolicy, map[securityBean.Severity]*CvePolicy) {
	cvePolicy := make(map[string][]*CvePolicy)
	severityPolicy := make(map[securityBean.Severity][]*CvePolicy)
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

func (impl *CvePolicyRepositoryImpl) getHighestPolicyS(allPolicies map[securityBean.Severity][]*CvePolicy) map[securityBean.Severity]*CvePolicy {
	applicablePolicies := make(map[securityBean.Severity]*CvePolicy)
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

// GetActiveByCveIdAndScope returns cvePolicy slice, based on scope, if all scopes are 0 or null then global cvePolicies are returned
func (impl *CvePolicyRepositoryImpl) GetActiveByCveIdAndScope(cveId string, envId, appId, clusterId int) (policies []*CvePolicy, err error) {
	query := impl.dbConnection.Model(&policies).
		Column("cve_policy.*").
		Where("deleted = ?", false).
		Where("cve_store_id = ?", cveId)
	if appId == 0 && envId > 0 {
		query = query.Where("env_id = ?", envId)
	}
	if clusterId > 0 {
		query = query.Where("cluster_id = ?", clusterId)
	}
	if appId > 0 && envId > 0 {
		query.WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q = q.WhereOrGroup(func(sq *orm.Query) (*orm.Query, error) {
				sq = sq.Where("app_id = ?", appId).Where("env_id = ?", envId)
				return sq, nil
			}).
				WhereOrGroup(func(sq *orm.Query) (*orm.Query, error) {
					sq = sq.Where("app_id = ?", appId).Where("env_id is null")
					return sq, nil
				})
			return q, nil
		})
	} else if appId > 0 && envId == 0 {
		query = query.Where("app_id = ?", appId)
	}
	err = query.Select()
	if err != nil {
		impl.logger.Errorw("error in getting active cvePolicies by cveId", "cveId", cveId, "err", err)
		return nil, err
	}
	return policies, nil
}

func (impl *CvePolicyRepositoryImpl) UpdatePoliciesInBulk(tx *pg.Tx, policies []*CvePolicy) error {
	if tx != nil {
		_, err := tx.Model(&policies).Update()
		if err != nil {
			impl.logger.Errorw("error in updating policies in transaction", "err", err)
			return err
		}
	} else {
		_, err := impl.dbConnection.Model(&policies).Update()
		if err != nil {
			impl.logger.Errorw("error in updating policies", "err", err)
			return err
		}
	}
	return nil
}
