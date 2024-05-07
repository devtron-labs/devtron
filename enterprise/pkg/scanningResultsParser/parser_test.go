package scanningResultsParser

import (
	"reflect"
	"testing"
)

func Test_parseLicense(t *testing.T) {
	type args struct {
		scanResult string
	}
	tests := []struct {
		name string
		args args
		want *Licenses
	}{{name: "tets1", args: args{scanResult: testData}}}// TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseLicense(tt.args.scanResult); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLicense() = %v, want %v", got, tt.want)
			}
		})
	}
}

const testData = `{
  "SchemaVersion": 2,
  "ArtifactName": "/security/devtronimagescan/344/code",
  "ArtifactType": "filesystem",
  "Metadata": {
    "ImageConfig": {
      "architecture": "",
      "created": "0001-01-01T00:00:00Z",
      "os": "",
      "rootfs": {
        "type": "",
        "diff_ids": null
      },
      "config": {}
    }
  },
  "Results": [
    {
      "Target": "go.mod",
      "Class": "lang-pkgs",
      "Type": "gomod",
      "Vulnerabilities": [
        {
          "VulnerabilityID": "CVE-2024-24557",
          "PkgID": "github.com/docker/docker@v20.10.27+incompatible",
          "PkgName": "github.com/docker/docker",
          "InstalledVersion": "20.10.27+incompatible",
          "FixedVersion": "25.0.2, 24.0.9",
          "Status": "fixed",
          "Layer": {},
          "SeveritySource": "ghsa",
          "PrimaryURL": "https://avd.aquasec.com/nvd/cve-2024-24557",
          "DataSource": {
            "ID": "ghsa",
            "Name": "GitHub Security Advisory Go",
            "URL": "https://github.com/advisories?query=type%3Areviewed+ecosystem%3Ago"
          },
          "Title": "moby: classic builder cache poisoning",
          "Description": "Moby is an open-source project created by Docker to enable software containerization. The classic builder cache system is prone to cache poisoning if the image is built FROM scratch. Also, changes to some instructions (most important being HEALTHCHECK and ONBUILD) would not cause a cache miss. An attacker with the knowledge of the Dockerfile someone is using could poison their cache by making them pull a specially crafted image that would be considered as a valid cache candidate for some build steps. 23.0+ users are only affected if they explicitly opted out of Buildkit (DOCKER_BUILDKIT=0 environment variable) or are using the /build API endpoint. All users on versions older than 23.0 could be impacted. Image build API endpoint (/build) and ImageBuild function from github.com/docker/docker/client is also affected as it the uses classic builder by default. Patches are included in 24.0.9 and 25.0.2 releases.",
          "Severity": "MEDIUM",
          "CweIDs": [
            "CWE-346",
            "CWE-345"
          ],
          "CVSS": {
            "ghsa": {
              "V3Vector": "CVSS:3.1/AV:L/AC:H/PR:N/UI:R/S:C/C:L/I:H/A:L",
              "V3Score": 6.9
            },
            "nvd": {
              "V3Vector": "CVSS:3.1/AV:L/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H",
              "V3Score": 7.8
            },
            "redhat": {
              "V3Vector": "CVSS:3.1/AV:L/AC:H/PR:N/UI:R/S:C/C:L/I:H/A:L",
              "V3Score": 6.9
            }
          },
          "References": [
            "https://access.redhat.com/security/cve/CVE-2024-24557",
            "https://github.com/moby/moby",
            "https://github.com/moby/moby/commit/3e230cfdcc989dc524882f6579f9e0dac77400ae",
            "https://github.com/moby/moby/commit/fca702de7f71362c8d103073c7e4a1d0a467fadd",
            "https://github.com/moby/moby/commit/fce6e0ca9bc000888de3daa157af14fa41fcd0ff",
            "https://github.com/moby/moby/security/advisories/GHSA-xw73-rw38-6vjc",
            "https://nvd.nist.gov/vuln/detail/CVE-2024-24557",
            "https://www.cve.org/CVERecord?id=CVE-2024-24557"
          ],
          "PublishedDate": "2024-02-01T17:15:10.953Z",
          "LastModifiedDate": "2024-02-09T20:21:32.97Z"
        },
        {
          "VulnerabilityID": "CVE-2024-29018",
          "PkgID": "github.com/docker/docker@v20.10.27+incompatible",
          "PkgName": "github.com/docker/docker",
          "InstalledVersion": "20.10.27+incompatible",
          "FixedVersion": "26.0.0-rc3, 25.0.5, 23.0.11",
          "Status": "fixed",
          "Layer": {},
          "SeveritySource": "ghsa",
          "PrimaryURL": "https://avd.aquasec.com/nvd/cve-2024-29018",
          "DataSource": {
            "ID": "ghsa",
            "Name": "GitHub Security Advisory Go",
            "URL": "https://github.com/advisories?query=type%3Areviewed+ecosystem%3Ago"
          },
          "Title": "moby: external DNS requests from 'internal' networks could lead to data exfiltration",
          "Description": "Moby is an open source container framework that is a key component of Docker Engine, Docker Desktop, and other distributions of container tooling or runtimes. Moby's networking implementation allows for many networks, each with their own IP address range and gateway, to be defined. This feature is frequently referred to as custom networks, as each network can have a different driver, set of parameters and thus behaviors. When creating a network, the \--internal flag is used to designate a network as _internal_. The internal attribute in a docker-compose.yml file may also be used to mark a network _internal_, and other API clients may specify the internal parameter as well.\n\nWhen containers with networking are created, they are assigned unique network interfaces and IP addresses. The host serves as a router for non-internal networks, with a gateway IP that provides SNAT/DNAT to/from container IPs.\n\nContainers on an internal network may communicate between each other, but are precluded from communicating with any networks the host has access to (LAN or WAN) as no default route is configured, and firewall rules are set up to drop all outgoing traffic. Communication with the gateway IP address (and thus appropriately configured host services) is possible, and the host may communicate with any container IP directly.\n\nIn addition to configuring the Linux kernel's various networking features to enable container networking, ,dockerd, directly provides some services to container networks. Principal among these is serving as a resolver, enabling service discovery, and resolution of names from an upstream resolver.\n\nWhen a DNS request for a name that does not correspond to a container is received, the request is forwarded to the configured upstream resolver. This request is made from the container's network namespace: the level of access and routing of traffic is the same as if the request was made by the container itself.\n\nAs a consequence of this design, containers solely attached to an internal network will be unable to resolve names using the upstream resolver, as the container itself is unable to communicate with that nameserver. Only the names of containers also attached to the internal network are able to be resolved.\n\nMany systems run a local forwarding DNS resolver. As the host and any containers have separate loopback devices, a consequence of the design described above is that containers are unable to resolve names from the host's configured resolver, as they cannot reach these addresses on the host loopback device. To bridge this gap, and to allow containers to properly resolve names even when a local forwarding resolver is used on a loopback address, ,dockerd, detects this scenario and instead forward DNS requests from the host namework namespace. The loopback resolver then forwards the requests to its configured upstream resolvers, as expected.\n\nBecause ,dockerd, forwards DNS requests to the host loopback device, bypassing the container network namespace's normal routing semantics entirely, internal networks can unexpectedly forward DNS requests to an external nameserver. By registering a domain for which they control the authoritative nameservers, an attacker could arrange for a compromised container to exfiltrate data by encoding it in DNS queries that will eventually be answered by their nameservers.\n\nDocker Desktop is not affected, as Docker Desktop always runs an internal resolver on a RFC 1918 address.\n\nMoby releases 26.0.0, 25.0.4, and 23.0.11 are patched to prevent forwarding any DNS requests from internal networks. As a workaround, run containers intended to be solely attached to internal networks with a custom upstream address, which will force all upstream DNS queries to be resolved from the container's network namespace.",
          "Severity": "MEDIUM",
          "CweIDs": [
            "CWE-669"
          ],
          "CVSS": {
            "ghsa": {
              "V3Vector": "CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:U/C:H/I:N/A:N",
              "V3Score": 5.9
            },
            "redhat": {
              "V3Vector": "CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:U/C:H/I:N/A:N",
              "V3Score": 5.9
            }
          },
          "References": [
            "https://access.redhat.com/security/cve/CVE-2024-29018",
            "https://github.com/moby/moby",
            "https://github.com/moby/moby/pull/46609",
            "https://github.com/moby/moby/security/advisories/GHSA-mq39-4gv4-mvpx",
            "https://nvd.nist.gov/vuln/detail/CVE-2024-29018",
            "https://www.cve.org/CVERecord?id=CVE-2024-29018"
          ],
          "PublishedDate": "2024-03-20T21:15:31.113Z",
          "LastModifiedDate": "2024-03-21T12:58:51.093Z"
        }
      ]
    },
    {
      "Target": "Dockerfile",
      "Class": "config",
      "Type": "dockerfile",
      "MisconfSummary": {
        "Successes": 23,
        "Failures": 3,
        "Exceptions": 0
      },
      "Misconfigurations": [
        {
          "Type": "Dockerfile Security Check",
          "ID": "DS005",
          "AVDID": "AVD-DS-0005",
          "Title": "ADD instead of COPY",
          "Description": "You should use COPY instead of ADD unless you want to extract a tar file. Note that an ADD command will extract a tar file, which adds the risk of Zip-based vulnerabilities. Accordingly, it is advised to use a COPY command, which does not extract tar files.",
          "Message": "Consider using 'COPY . /go/src/github.com/devtron-labs/image-scanner' command instead of 'ADD . /go/src/github.com/devtron-labs/image-scanner'",
          "Namespace": "builtin.dockerfile.DS005",
          "Query": "data.builtin.dockerfile.DS005.deny",
          "Resolution": "Use COPY instead of ADD",
          "Severity": "LOW",
          "PrimaryURL": "https://avd.aquasec.com/misconfig/ds005",
          "References": [
            "https://docs.docker.com/engine/reference/builder/#add",
            "https://avd.aquasec.com/misconfig/ds005"
          ],
          "Status": "FAIL",
          "Layer": {},
          "CauseMetadata": {
            "Provider": "Dockerfile",
            "Service": "general",
            "StartLine": 6,
            "EndLine": 6,
            "Code": {
              "Lines": [
                {
                  "Number": 6,
                  "Content": "ADD . /go/src/github.com/devtron-labs/image-scanner",
                  "IsCause": true,
                  "Annotation": "",
                  "Truncated": false,
                  "FirstCause": true,
                  "LastCause": true
                }
              ]
            }
          }
        },
        {
          "Type": "Dockerfile Security Check",
          "ID": "DS025",
          "AVDID": "AVD-DS-0025",
          "Title": "'apk add' is missing '--no-cache'",
          "Description": "You should use 'apk add' with '--no-cache' to clean package cached data and reduce image size.",
          "Message": "'--no-cache' is missed: apk add --update make",
          "Namespace": "builtin.dockerfile.DS025",
          "Query": "data.builtin.dockerfile.DS025.deny",
          "Resolution": "Add '--no-cache' to 'apk add' in Dockerfile",
          "Severity": "HIGH",
          "PrimaryURL": "https://avd.aquasec.com/misconfig/ds025",
          "References": [
            "https://github.com/gliderlabs/docker-alpine/blob/master/docs/usage.md#disabling-cache",
            "https://avd.aquasec.com/misconfig/ds025"
          ],
          "Status": "FAIL",
          "Layer": {},
          "CauseMetadata": {
            "Provider": "Dockerfile",
            "Service": "general",
            "StartLine": 3,
            "EndLine": 3,
            "Code": {
              "Lines": [
                {
                  "Number": 3,
                  "Content": "RUN apk add --update make",
                  "IsCause": true,
                  "Annotation": "",
                  "Truncated": false,
                  "FirstCause": true,
                  "LastCause": true
                }
              ]
            }
          }
        },
        {
          "Type": "Dockerfile Security Check",
          "ID": "DS026",
          "AVDID": "AVD-DS-0026",
          "Title": "No HEALTHCHECK defined",
          "Description": "You should add HEALTHCHECK instruction in your docker container images to perform the health check on running containers.",
          "Message": "Add HEALTHCHECK instruction in your Dockerfile",
          "Namespace": "builtin.dockerfile.DS026",
          "Query": "data.builtin.dockerfile.DS026.deny",
          "Resolution": "Add HEALTHCHECK instruction in Dockerfile",
          "Severity": "LOW",
          "PrimaryURL": "https://avd.aquasec.com/misconfig/ds026",
          "References": [
            "https://blog.aquasec.com/docker-security-best-practices",
            "https://avd.aquasec.com/misconfig/ds026"
          ],
          "Status": "FAIL",
          "Layer": {},
          "CauseMetadata": {
            "Provider": "Dockerfile",
            "Service": "general",
            "Code": {
              "Lines": null
            }
          }
        }
      ]
    },
    {
      "Target": "OS Packages",
      "Class": "license"
    },
    {
      "Target": "go.mod",
      "Class": "license"
    },
    {
      "Target": "Loose File License(s)",
      "Class": "license-file",
      "Licenses": [
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/mellium.im/sasl/LICENSE",
          "Name": "BSD-2-Clause",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/BSD-2-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/go-pg/pg/LICENSE",
          "Name": "BSD-2-Clause",
          "Confidence": 0.994535519125683,
          "Link": "https://spdx.org/licenses/BSD-2-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/matttproud/golang_protobuf_extensions/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/coreos/clair/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/optiopay/klar/LICENSE",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/jmespath/go-jmespath/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 0.9285714285714286,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/arl/statsviz/LICENSE",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/beorn7/perks/LICENSE",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/cespare/xxhash/v2/LICENSE.txt",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/docker/cli/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/docker/distribution/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/docker/docker-credential-helpers/LICENSE",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/docker/docker/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/google/go-containerregistry/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/google/wire/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/quay/claircore/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/antihax/optional/LICENSE",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/aws/aws-sdk-go/LICENSE.txt",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/aws/aws-sdk-go/internal/sync/singleflight/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/tidwall/pretty/LICENSE",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/tidwall/gjson/LICENSE",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/tidwall/match/LICENSE",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/caarlos0/env/LICENSE.md",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/caarlos0/env/v6/LICENSE.md",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/klauspost/compress/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 0.9961783439490446,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/klauspost/compress/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/klauspost/compress/LICENSE",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/nats-io/nats.go/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/nats-io/nkeys/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/nats-io/nuid/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/pkg/errors/LICENSE",
          "Name": "BSD-2-Clause",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/BSD-2-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/devtron-labs/common-lib/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/golang/mock/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/golang/protobuf/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "HIGH",
          "Category": "restricted",
          "PkgName": "",
          "FilePath": "vendor/github.com/juju/errors/LICENSE",
          "Name": "LGPL-3.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/LGPL-3.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/prometheus/client_golang/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/prometheus/client_model/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/prometheus/common/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/prometheus/common/internal/bitbucket.org/ww/goautoneg/README.txt",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9767441860465116,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/prometheus/procfs/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/grpc-ecosystem/grpc-gateway/LICENSE.txt",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9906976744186047,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/jinzhu/inflection/LICENSE",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/opencontainers/go-digest/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/opencontainers/image-spec/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/Knetic/govaluate/LICENSE",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/gorilla/websocket/LICENSE",
          "Name": "BSD-2-Clause",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/BSD-2-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/gorilla/mux/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/github.com/sirupsen/logrus/LICENSE",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/google.golang.org/appengine/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/google.golang.org/genproto/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/google.golang.org/genproto/googleapis/api/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/google.golang.org/genproto/googleapis/rpc/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/google.golang.org/protobuf/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/google.golang.org/grpc/NOTICE.txt",
          "Name": "Apache-2.0",
          "Confidence": 0.9285714285714286,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/google.golang.org/grpc/regenerate.sh",
          "Name": "Apache-2.0",
          "Confidence": 0.9285714285714286,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/google.golang.org/grpc/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/cloud.google.com/go/compute/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/cloud.google.com/go/compute/metadata/LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/go.uber.org/atomic/LICENSE.txt",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/go.uber.org/multierr/LICENSE.txt",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/go.uber.org/zap/LICENSE.txt",
          "Name": "MIT",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/MIT.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/golang.org/x/crypto/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/golang.org/x/mod/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/golang.org/x/net/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/golang.org/x/oauth2/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/golang.org/x/sync/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/golang.org/x/sys/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/golang.org/x/text/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "vendor/golang.org/x/tools/LICENSE",
          "Name": "BSD-3-Clause",
          "Confidence": 0.9812206572769953,
          "Link": "https://spdx.org/licenses/BSD-3-Clause.html"
        },
        {
          "Severity": "LOW",
          "Category": "notice",
          "PkgName": "",
          "FilePath": "LICENSE",
          "Name": "Apache-2.0",
          "Confidence": 1,
          "Link": "https://spdx.org/licenses/Apache-2.0.html"
        }
      ]
    }
  ]
}
`
