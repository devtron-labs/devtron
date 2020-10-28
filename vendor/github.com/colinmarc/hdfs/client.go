package hdfs

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"strings"

	hdfs "github.com/colinmarc/hdfs/protocol/hadoop_hdfs"
	"github.com/colinmarc/hdfs/rpc"
	krb "gopkg.in/jcmturner/gokrb5.v5/client"
)

// A Client represents a connection to an HDFS cluster
type Client struct {
	namenode *rpc.NamenodeConnection
	defaults *hdfs.FsServerDefaultsProto
	options  ClientOptions
}

// ClientOptions represents the configurable options for a client.
// The NamenodeDialFunc and DatanodeDialFunc options can be used to set
// connection timeouts:
//
//    dialFunc := (&net.Dialer{
//        Timeout:   30 * time.Second,
//        KeepAlive: 30 * time.Second,
//        DualStack: true,
//    }).DialContext
//
//    options := ClientOptions{
//        Addresses: []string{"nn1:9000"},
//        NamenodeDialFunc: dialFunc,
//        DatanodeDialFunc: dialFunc,
//    }
type ClientOptions struct {
	// Addresses specifies the namenode(s) to connect to.
	Addresses []string
	// User specifies which HDFS user the client will act as. It is required
	// unless kerberos authentication is enabled, in which case it will be
	// determined from the provided credentials if empty.
	User string
	// UseDatanodeHostname specifies whether the client should connect to the
	// datanodes via hostname (which is useful in multi-homed setups) or IP
	// address, which may be required if DNS isn't available.
	UseDatanodeHostname bool
	// NamenodeDialFunc is used to connect to the datanodes. If nil, then
	// (&net.Dialer{}).DialContext is used.
	NamenodeDialFunc func(ctx context.Context, network, addr string) (net.Conn, error)
	// DatanodeDialFunc is used to connect to the datanodes. If nil, then
	// (&net.Dialer{}).DialContext is used.
	DatanodeDialFunc func(ctx context.Context, network, addr string) (net.Conn, error)
	// KerberosClient is used to connect to kerberized HDFS clusters. If provided,
	// the client will always mutually athenticate when connecting to the
	// namenode(s).
	KerberosClient *krb.Client
	// KerberosServicePrincipleName specifies the Service Principle Name
	// (<SERVICE>/<FQDN>) for the namenode(s). Like in the
	// dfs.namenode.kerberos.principal property of core-site.xml, the special
	// string '_HOST' can be substituted for the address of the namenode in a
	// multi-namenode setup (for example: 'nn/_HOST'). It is required if
	// KerberosClient is provided.
	KerberosServicePrincipleName string
	// Namenode optionally specifies an existing NamenodeConnection to wrap. This
	// is useful if you needed to create the namenode net.Conn manually for
	// whatever reason.
	//
	// Deprecated: use NamenodeDialFunc instead.
	Namenode *rpc.NamenodeConnection
}

// ClientOptionsFromConf attempts to load any relevant configuration options
// from the given Hadoop configuration and create a ClientOptions struct
// suitable for creating a Client. Currently this sets the following fields
// on the resulting ClientOptions:
//
//   // Determined by fs.defaultFS (or the deprecated fs.default.name), or
//   // fields beginning with dfs.namenode.rpc-address.
//   Addresses []string
//
//   // Determined by dfs.client.use.datanode.hostname.
//   UseDatanodeHostname bool
//
//   // Set to a non-nil but empty client (without credentials) if the value of
//   // hadoop.security.authentication is 'kerberos'. It must then be replaced
//   // with a credentialed Kerberos client.
//   KerberosClient *krb.Client
//
//   // Determined by dfs.namenode.kerberos.principal, with the realm
//   // (everything after the first '@') chopped off.
//   KerberosServicePrincipleName string
//
// Because of the way Kerberos can be forced by the Hadoop configuration but not
// actually configured, you should check for whether KerberosClient is set in
// the resulting ClientOptions before proceeding:
//
//   options, _ := ClientOptionsFromConf(conf)
//   if options.KerberosClient != nil {
//      // Replace with a valid credentialed client.
//      options.KerberosClient = getKerberosClient()
//   }
func ClientOptionsFromConf(conf HadoopConf) (ClientOptions, error) {
	namenodes, err := conf.Namenodes()
	options := ClientOptions{Addresses: namenodes}

	options.UseDatanodeHostname = (conf["dfs.client.use.datanode.hostname"] == "true")

	if strings.ToLower(conf["hadoop.security.authentication"]) == "kerberos" {
		// Set an empty KerberosClient here so that the user is forced to either
		// unset it (disabling kerberos altogether) or replace it with a valid
		// client. If the user does neither, NewClient will return an error.
		options.KerberosClient = &krb.Client{}
	}

	if conf["dfs.namenode.kerberos.principal"] != "" {
		options.KerberosServicePrincipleName = strings.Split(conf["dfs.namenode.kerberos.principal"], "@")[0]
	}

	return options, err
}

// NewClient returns a connected Client for the given options, or an error if
// the client could not be created.
func NewClient(options ClientOptions) (*Client, error) {
	var err error
	if options.Namenode == nil {
		if options.KerberosClient != nil && options.KerberosClient.Credentials == nil {
			return nil, errors.New("kerberos enabled, but kerberos client is missing credentials")
		}

		if options.KerberosClient != nil && options.KerberosServicePrincipleName == "" {
			return nil, errors.New("kerberos enabled, but kerberos namenode SPN is not provided")
		}

		if options.User == "" {
			if options.KerberosClient != nil {
				creds := options.KerberosClient.Credentials
				options.User = creds.Username + "@" + creds.Realm
			} else {
				return nil, errors.New("user not specified")
			}
		}

		options.Namenode, err = rpc.NewNamenodeConnectionWithOptions(
			rpc.NamenodeConnectionOptions{
				Addresses:                    options.Addresses,
				User:                         options.User,
				DialFunc:                     options.NamenodeDialFunc,
				KerberosClient:               options.KerberosClient,
				KerberosServicePrincipleName: options.KerberosServicePrincipleName,
			},
		)

		if err != nil {
			return nil, err
		}
	}

	return &Client{namenode: options.Namenode, options: options}, nil
}

// New returns a connected Client, or an error if it can't connect. The user
// will be the current system user. Any relevantoptions (including the
// address(es) of the namenode(s), if an empty string is passed) will be loaded
// from the Hadoop configuration present at HADOOP_CONF_DIR.  Note, however,
// that New will not attempt any Kerberos authentication; use NewClient if you
// need that.
func New(address string) (*Client, error) {
	conf := LoadHadoopConf("")
	options, err := ClientOptionsFromConf(conf)
	if err != nil {
		options = ClientOptions{}
	}

	if address != "" {
		options.Addresses = strings.Split(address, ",")
	}

	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	options.User = u.Username
	return NewClient(options)
}

// NewForUser returns a connected Client with the user specified, or an error if
// it can't connect.
//
// Deprecated: Use NewClient with ClientOptions instead.
func NewForUser(address string, user string) (*Client, error) {
	return NewClient(ClientOptions{
		Addresses: []string{address},
		User:      user,
	})
}

// NewForConnection returns Client with the specified, underlying rpc.NamenodeConnection.
// You can use rpc.WrapNamenodeConnection to wrap your own net.Conn.
//
// Deprecated: Use NewClient with ClientOptions instead.
func NewForConnection(namenode *rpc.NamenodeConnection) *Client {
	client, _ := NewClient(ClientOptions{Namenode: namenode})
	return client
}

// ReadFile reads the file named by filename and returns the contents.
func (c *Client) ReadFile(filename string) ([]byte, error) {
	f, err := c.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return ioutil.ReadAll(f)
}

// CopyToLocal copies the HDFS file specified by src to the local file at dst.
// If dst already exists, it will be overwritten.
func (c *Client) CopyToLocal(src string, dst string) error {
	remote, err := c.Open(src)
	if err != nil {
		return err
	}
	defer remote.Close()

	local, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer local.Close()

	_, err = io.Copy(local, remote)
	return err
}

// CopyToRemote copies the local file specified by src to the HDFS file at dst.
func (c *Client) CopyToRemote(src string, dst string) error {
	local, err := os.Open(src)
	if err != nil {
		return err
	}
	defer local.Close()

	remote, err := c.Create(dst)
	if err != nil {
		return err
	}
	defer remote.Close()

	_, err = io.Copy(remote, local)
	return err
}

func (c *Client) fetchDefaults() (*hdfs.FsServerDefaultsProto, error) {
	if c.defaults != nil {
		return c.defaults, nil
	}

	req := &hdfs.GetServerDefaultsRequestProto{}
	resp := &hdfs.GetServerDefaultsResponseProto{}

	err := c.namenode.Execute("getServerDefaults", req, resp)
	if err != nil {
		return nil, err
	}

	c.defaults = resp.GetServerDefaults()
	return c.defaults, nil
}

// Close terminates all underlying socket connections to remote server.
func (c *Client) Close() error {
	return c.namenode.Close()
}

// Username returns the current system user if it is not set.
//
// Deprecated: just use user.Current. Previous versions of this function would
// check the env variable HADOOP_USER_NAME; this functionality was removed.
func Username() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	return currentUser.Username, nil
}
