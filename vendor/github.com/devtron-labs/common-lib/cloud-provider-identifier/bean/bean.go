package bean

const (
	Amazon       = "amazon"
	Alibaba      = "alibaba"
	Azure        = "azure"
	Google       = "google"
	Oracle       = "oracle"
	DigitalOcean = "digitalocean"
	Unknown      = "unknown"
)

const (
	AlibabaIdentifierString      = "Alibaba Cloud"
	AmazonIdentifierString       = "amazon"
	AzureIdentifierString        = "Microsoft Corporation"
	DigitalOceanIdentifierString = "DigitalOcean"
	GoogleIdentifierString       = "Google"
	OracleIdentifierString       = "OracleCloud"
)

const (
	AlibabaSysFile      = "/sys/class/dmi/id/product_name"
	AmazonSysFile       = "/sys/class/dmi/id/product_version"
	AzureSysFile        = "/sys/class/dmi/id/sys_vendor"
	DigitalOceanSysFile = "/sys/class/dmi/id/sys_vendor"
	GoogleSysFile       = "/sys/class/dmi/id/product_name"
	OracleSysFile       = "/sys/class/dmi/id/chassis_asset_tag"
)

const (
	AlibabaMetadataServer      = "http://100.100.100.200/latest/meta-data/instance/instance-type"
	AmazonMetadataServer       = "http://169.254.169.254/latest/dynamic/instance-identity/document"
	AzureMetadataServer        = "http://169.254.169.254/metadata/instance?api-version=2017-12-01"
	DigitalOceanMetadataServer = "http://169.254.169.254/metadata/v1.json"
	GoogleMetadataServer       = "http://metadata.google.internal/computeMetadata/v1/instance/tags"
	OracleMetadataServer       = "http://169.254.169.254/opc/v1/instance/metadata/"
)
