#!/bin/sh

set -e

KERBEROS=${KERBEROS-"false"}

UBUNTU_CODENAME=$(lsb_release -c | awk '{print $2}')

sudo tee /etc/apt/sources.list.d/cdh.list <<EOF
deb [arch=amd64] http://archive.cloudera.com/cdh5/ubuntu/$UBUNTU_CODENAME/amd64/cdh $UBUNTU_CODENAME-cdh5 contrib
EOF

sudo tee /etc/apt/preferences.d/cloudera.pref <<EOF
Package: *
Pin: release o=Cloudera, l=Cloudera
Pin-Priority: 501
EOF

sudo apt-get update

CONF_AUTHENTICATION="simple"
if [ $KERBEROS = "true" ]; then
  CONF_AUTHENTICATION="kerberos"

  HOSTNAME=$(hostname)

  KERBEROS_REALM="EXAMPLE.COM"
  KERBEROS_PRINCIPLE="administrator"
  KERBEROS_PASSWORD="password1234"

  sudo tee /etc/krb5.conf << EOF
[libdefaults]
    default_realm = $KERBEROS_REALM
    dns_lookup_realm = false
    dns_lookup_kdc = false
[realms]
    $KERBEROS_REALM = {
        kdc = localhost
        admin_server = localhost
    }
[logging]
    default = FILE:/var/log/krb5libs.log
    kdc = FILE:/var/log/krb5kdc.log
    admin_server = FILE:/var/log/kadmind.log
[domain_realm]
    .localhost = $KERBEROS_REALM
    localhost = $KERBEROS_REALM
EOF

  sudo mkdir /etc/krb5kdc
  sudo printf '*/*@%s\t*' "$KERBEROS_REALM" | sudo tee /etc/krb5kdc/kadm5.acl

  sudo apt-get install -y krb5-user krb5-kdc krb5-admin-server

  printf "$KERBEROS_PASSWORD\n$KERBEROS_PASSWORD" | sudo kdb5_util -r "$KERBEROS_REALM" create -s
  for p in nn dn travis gohdfs1 gohdfs2; do
    sudo kadmin.local -q "addprinc -randkey $p/$HOSTNAME@$KERBEROS_REALM"
    sudo kadmin.local -q "addprinc -randkey $p/localhost@$KERBEROS_REALM"
    sudo kadmin.local -q "xst -k /tmp/$p.keytab $p/$HOSTNAME@$KERBEROS_REALM"
    sudo kadmin.local -q "xst -k /tmp/$p.keytab $p/localhost@$KERBEROS_REALM"
    sudo chmod +rx /tmp/$p.keytab
  done

  sudo service krb5-kdc restart
  sudo service krb5-admin-server restart

  kinit -kt /tmp/travis.keytab "travis/localhost@$KERBEROS_REALM"

  # The go tests need ccache files for these principles in a specific place.
  for p in travis gohdfs1 gohdfs2; do
    kinit -kt "/tmp/$p.keytab" -c "/tmp/krb5cc_gohdfs_$p" "$p/localhost@$KERBEROS_REALM"
  done
fi

sudo mkdir -p /etc/hadoop/conf.gohdfs
sudo tee /etc/hadoop/conf.gohdfs/core-site.xml <<EOF
<configuration>
  <property>
    <name>fs.defaultFS</name>
    <value>hdfs://localhost:9000</value>
  </property>
  <property>
    <name>hadoop.security.authentication</name>
    <value>$CONF_AUTHENTICATION</value>
  </property>
  <property>
    <name>hadoop.security.authorization</name>
    <value>$KERBEROS</value>
  </property>
  <property>
    <name>dfs.namenode.keytab.file</name>
    <value>/tmp/nn.keytab</value>
  </property>
  <property>
    <name>dfs.namenode.kerberos.principal</name>
    <value>nn/localhost@$KERBEROS_REALM</value>
  </property>
  <property>
    <name>dfs.web.authentication.kerberos.principal</name>
    <value>nn/localhost@$KERBEROS_REALM</value>
  </property>
  <property>
    <name>dfs.datanode.keytab.file</name>
    <value>/tmp/dn.keytab</value>
  </property>
  <property>
    <name>dfs.datanode.kerberos.principal</name>
    <value>dn/localhost@$KERBEROS_REALM</value>
  </property>
</configuration>
EOF

sudo tee /etc/hadoop/conf.gohdfs/hdfs-site.xml <<EOF
<configuration>
  <property>
    <name>dfs.namenode.name.dir</name>
    <value>/opt/hdfs/name</value>
  </property>
  <property>
    <name>dfs.datanode.data.dir</name>
    <value>/opt/hdfs/data</value>
  </property>
  <property>
   <name>dfs.permissions.superusergroup</name>
   <value>hadoop</value>
  </property>
  <property>
    <name>dfs.safemode.extension</name>
    <value>0</value>
  </property>
  <property>
     <name>dfs.safemode.min.datanodes</name>
     <value>1</value>
  </property>
  <property>
    <name>dfs.block.access.token.enable</name>
    <value>$KERBEROS</value>
  </property>
  <property>
    <name>ignore.secure.ports.for.testing</name>
    <value>true</value>
  </property>
</configuration>
EOF

sudo update-alternatives --install /etc/hadoop/conf hadoop-conf /etc/hadoop/conf.gohdfs 99
sudo apt-get install -y --allow-unauthenticated hadoop-hdfs-namenode hadoop-hdfs-datanode

sudo mkdir -p /opt/hdfs/data /opt/hdfs/name
sudo chown -R hdfs:hdfs /opt/hdfs
sudo -u hdfs hdfs namenode -format -nonInteractive

sudo adduser travis hadoop

sudo service hadoop-hdfs-datanode restart
sudo service hadoop-hdfs-namenode restart

hdfs dfsadmin -safemode wait
