package client

type Subnet struct {
	ID                  int64  `json:"id,omitempty"`
	CIDR                string `json:"cidr,omitempty"`
	Name                string `json:"name,omitempty"`
	IPVersion           int64  `json:"ip_version,omitempty"`
	VLANID              *int64 `json:"vlan_id,omitempty"`
	Description         string `json:"description,omitempty"`
	Notes               string `json:"notes,omitempty"`
	ParentID            *int64 `json:"parent_id,omitempty"`
	ScanIntervalMinutes *int64 `json:"scan_interval_minutes,omitempty"`
	DNSProviderName     string `json:"dns_provider_name,omitempty"`
	DHCPProviderName    string `json:"dhcp_provider_name,omitempty"`
	RequestEligible     bool   `json:"request_eligible,omitempty"`
}

type Address struct {
	ID          int64  `json:"id,omitempty"`
	Address     string `json:"address,omitempty"`
	SubnetID    int64  `json:"subnet_id,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
	Status      string `json:"status,omitempty"`
	MACAddress  string `json:"mac_address,omitempty"`
	Description string `json:"description,omitempty"`
	Notes       string `json:"notes,omitempty"`
	LastSeen    string `json:"last_seen,omitempty"`
}

type AllocateRequest struct {
	Hostname     string `json:"hostname"`
	Description  string `json:"description,omitempty"`
	MACAddress   string `json:"mac_address,omitempty"`
	RegisterDNS  bool   `json:"register_dns"`
	RegisterDHCP bool   `json:"register_dhcp"`
	DNSZone      string `json:"dns_zone,omitempty"`
	RegisterPTR  bool   `json:"register_ptr,omitempty"`
}

type AllocateResult struct {
	ID             int64  `json:"id"`
	Address        string `json:"address"`
	SubnetID       int64  `json:"subnet_id"`
	SubnetCIDR     string `json:"subnet_cidr"`
	Hostname       string `json:"hostname"`
	Status         string `json:"status"`
	MACAddress     string `json:"mac_address"`
	DNSRegistered  bool   `json:"dns_registered"`
	DHCPRegistered bool   `json:"dhcp_registered"`
	PTRRegistered  bool   `json:"ptr_registered"`
	IsNew          bool   `json:"is_new"`
}

type Vlan struct {
	ID          int64  `json:"id,omitempty"`
	VLANID      int64  `json:"vlan_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

type DNSRecord struct {
	Name        string `json:"name"`
	RecordType  string `json:"record_type"`
	Value       string `json:"value"`
	Zone        string `json:"zone,omitempty"`
	TTL         int64  `json:"ttl,omitempty"`
	Source      string `json:"source,omitempty"`
	RegisterPTR bool   `json:"register_ptr,omitempty"`
}

type DHCPLease struct {
	ScopeID     string `json:"scope_id,omitempty"`
	IPAddress   string `json:"ip_address,omitempty"`
	MACAddress  string `json:"mac_address,omitempty"`
	ClientDUID  string `json:"client_duid,omitempty"`
	IAID        *int64 `json:"iaid,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Source      string `json:"source,omitempty"`
}

type deletePreview struct {
	Items []struct {
		Key string `json:"key"`
	} `json:"items"`
}

type pageEnvelope[T any] struct {
	Items  []T   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}
