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

package chartConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func getCr() ChartRepository {
	return nil
	//return NewChartRepository(models.GetDbConnection())
}

func TestChartRepositoryImpl_Save(t *testing.T) {
	cr := getCr()
	c := &Chart{
		ChartName:    "wordpress",
		ChartRepo:    "stable",
		ChartVersion: "5.1.1",
		Status:       models.CHARTSTATUS_NEW,
		Values:       `{"image":{"registry":"docker.io","repository":"bitnami/wordpress","tag":"5.0.3","pullPolicy":"IfNotPresent"},"wordpressUsername":"user","wordpressEmail":"user@example.com","wordpressFirstName":"FirstName","wordpressLastName":"LastName","wordpressBlogName":"User's Blog!","wordpressTablePrefix":"wp_","allowEmptyPassword":true,"replicaCount":1,"externalDatabase":{"host":"localhost","user":"bn_wordpress","password":"","database":"bitnami_wordpress","port":3306},"mariadb":{"enabled":true,"replication":{"enabled":false},"db":{"name":"bitnami_wordpress","user":"bn_wordpress"},"master":{"persistence":{"enabled":true,"accessMode":"ReadWriteOnce","size":"8Gi"}}},"service":{"type":"LoadBalancer","port":80,"httpsPort":443,"nodePorts":{"http":"","https":""},"externalTrafficPolicy":"Cluster","annotations":{}},"healthcheckHttps":false,"livenessProbe":{"initialDelaySeconds":120,"periodSeconds":10,"timeoutSeconds":5,"failureThreshold":6,"successThreshold":1},"readinessProbe":{"initialDelaySeconds":30,"periodSeconds":10,"timeoutSeconds":5,"failureThreshold":6,"successThreshold":1},"ingress":{"enabled":false,"certManager":false,"annotations":null,"hosts":[{"name":"wordpress.local","path":"/","tls":false,"tlsSecret":"wordpress.local-tls"}],"secrets":null},"persistence":{"enabled":true,"accessMode":"ReadWriteOnce","size":"10Gi"},"resources":{"requests":{"memory":"512Mi","cpu":"300m"}},"nodeSelector":{},"tolerations":[],"affinity":{},"podAnnotations":{},"metrics":{"enabled":false,"image":{"registry":"docker.io","repository":"lusotycoon/apache-exporter","tag":"v0.5.0","pullPolicy":"IfNotPresent"},"podAnnotations":{"prometheus.io/scrape":"true","prometheus.io/port":"9117"}}}`,
		AuditLog:     sql.AuditLog{CreatedBy: 1, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: 1},
	}

	err := cr.Save(c)
	assert.NoError(t, err)

}
