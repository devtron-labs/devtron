UPDATE "public"."attributes" SET value = 'quay.io/devtron/ubuntu-k8s-utils:latest,quay.io/devtron/alpine-k8s-utils:latest,quay.io/devtron/centos-k8s-utils:latest,quay.io/devtron/alpine-netshoot:latest',
                                 updated_on = NOW()
WHERE key = 'DEFAULT_TERMINAL_IMAGE_LIST';

ALTER table attributes alter column value TYPE character varying(250);