---same as 029800_021_xxxx.down.sql

UPDATE "public"."attributes"
SET value = '[{"groupId":"latest","groupRegex":"v1\\.2[4-8]\\..+","imageList":[{"image":"quay.io/devtron/ubuntu-k8s-utils:latest","name":"Ubuntu: Kubernetes utilites","description":"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on ubuntu OS"}, {"image":"quay.io/devtron/alpine-k8s-utils:latest","name":"Alpine: Kubernetes utilites","description":"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on alpine OS"},{"image":"quay.io/devtron/centos-k8s-utils:latest","name":"CentOS: Kubernetes utilites","description":"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on Cent OS"},{"image":"quay.io/devtron/alpine-netshoot:latest","name":"Alpine: Netshoot","description":"Contains Docker + Kubernetes network troubleshooting utilities."},{"image":"quay.io/devtron/k9s-k8s-utils:latest","name":"K9s: Kubernetes CLI","description": " Kubernetes CLI To Manage Your Clusters In Style!"}]} ,{"groupId":"v1.22","groupRegex":"v1\\.(21|22|23)\\..+","imageList":[{"image":"quay.io/devtron/ubuntu-k8s-utils:1.22","name":"Ubuntu: Kubernetes utilites","description":"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on ubuntu OS"},{"image":"quay.io/devtron/alpine-k8s-utils:1.22","name":"Alpine: Kubernetes utilites","description":"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on alpine OS"},{"image":"quay.io/devtron/centos-k8s-utils:1.22","name":"CentOS: Kubernetes utilites","description":"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on Cent OS"},{"image":"quay.io/devtron/alpine-netshoot:latest","name":"Alpine: Netshoot","description":"Contains Docker + Kubernetes network troubleshooting utilities."}]},{"groupId":"v1.19","groupRegex":"v1\\.(18|19|20)\\..+","imageList":[{"image":"quay.io/devtron/ubuntu-k8s-utils:1.19","name":"Ubuntu: Kubernetes utilites","description":"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on ubuntu OS"},{"image":"quay.io/devtron/alpine-k8s-utils:1.19","name":"Alpine: Kubernetes utilites","description":"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on alpine OS"},{"image":"quay.io/devtron/centos-k8s-utils:1.19","name":"CentOS: Kubernetes utilites","description":"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on Cent OS"},{"image":"quay.io/devtron/alpine-netshoot:latest","name":"Alpine: Netshoot","description":"Contains Docker + Kubernetes network troubleshooting utilities."},{"image":"quay.io/devtron/k9s-k8s-utils:latest","name":"K9s: Kubernetes CLI","description": " Kubernetes CLI To Manage Your Clusters In Style!"}]},{"groupId":"v1.16","groupRegex":"v1\\.(15|16|17)\\..+","imageList":[{"image":"quay.io/devtron/ubuntu-k8s-utils:1.16","name":"Ubuntu: Kubernetes utilites","description":"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on ubuntu OS"}, {"image":"quay.io/devtron/alpine-k8s-utils:1.16","name":"Alpine: Kubernetes utilites","description":"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on alpine OS"},{"image":"quay.io/devtron/centos-k8s-utils:1.16","name":"CentOS: Kubernetes utilites","description":"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on Cent OS"},{"image":"quay.io/devtron/alpine-netshoot:latest","name":"Alpine: Netshoot","description":"Contains Docker + Kubernetes network troubleshooting utilities."},{"image":"quay.io/devtron/k9s-k8s-utils:latest","name":"K9s: Kubernetes CLI","description": " Kubernetes CLI To Manage Your Clusters In Style!"}]}]',
    updated_on = NOW()
WHERE key = 'DEFAULT_TERMINAL_IMAGE_LIST';