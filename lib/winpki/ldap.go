/*
 * Teleport
 * Copyright (C) 2023  Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package winpki

import (
	"crypto/x509"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/go-ldap/ldap/v3"
	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/tlsca"
)

// LDAPConfig contains parameters for connecting to an LDAP server.
type LDAPConfig struct {
	// Addr is the LDAP server address in the form host:port.
	// Standard port is 636 for LDAPS.
	Addr string
	// Domain is an Active Directory domain name, like "example.com".
	Domain string
	// Username is an LDAP username, like "EXAMPLE\Administrator", where
	// "EXAMPLE" is the NetBIOS version of Domain.
	Username string
	// SID is the SID for the user specified by Username.
	SID string
	// InsecureSkipVerify decides whether we skip verifying with the LDAP server's CA when making the LDAPS connection.
	InsecureSkipVerify bool
	// ServerName is the name of the LDAP server for TLS.
	ServerName string
	// CA is an optional CA cert to be used for verification if InsecureSkipVerify is set to false.
	CA *x509.Certificate
}

// Check verifies this LDAPConfig
func (cfg LDAPConfig) Check() error {
	if cfg.Addr == "" {
		return trace.BadParameter("missing Addr in LDAPConfig")
	}
	if cfg.Domain == "" {
		return trace.BadParameter("missing Domain in LDAPConfig")
	}
	if cfg.Username == "" {
		return trace.BadParameter("missing Username in LDAPConfig")
	}
	return nil
}

// DomainDN returns the distinguished name for an Active Directory Domain.
func DomainDN(domain string) string {
	var sb strings.Builder
	parts := strings.SplitSeq(domain, ".")
	for p := range parts {
		if sb.Len() > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("DC=")
		sb.WriteString(p)
	}
	return sb.String()
}

const (
	// AttrObjectSid is the Security Identifier of an LDAP object
	AttrObjectSid = "objectSid"
	// AttrObjectClass is the object class of an LDAP object
	AttrObjectClass = "objectClass"
)

// classContainer is the object class for containers in Active Directory
const classContainer = "container"

// searchPageSize is desired page size for LDAP search. In Active Directory the default search size limit is 1000 entries,
// so in most cases the 1000 search page size will result in the optimal amount of requests made to
// LDAP server.
const searchPageSize = 1000

// LDAPClient is a windows LDAP client.
//
// It does not automatically detect when the underlying connection
// is closed. Callers should check for trace.ConnectionProblem errors
// and provide a new client with [SetClient].
type LDAPClient struct {
	mu     sync.Mutex
	client ldap.Client
}

// NewLDAPClient returns new LDAPClient. Parameter client may be nil.
func NewLDAPClient(client ldap.Client) *LDAPClient {
	return &LDAPClient{
		client: client,
	}
}

// SetClient sets the underlying ldap.Client
func (c *LDAPClient) SetClient(client ldap.Client) {
	c.mu.Lock()
	if c.client != nil {
		c.client.Close()
	}
	c.client = client
	c.mu.Unlock()
}

// Close closes the underlying ldap.Client
func (c *LDAPClient) Close() {
	c.mu.Lock()
	if c.client != nil {
		c.client.Close()
	}
	c.mu.Unlock()
}

// convertLDAPError attempts to convert LDAP error codes to their
// equivalent trace errors.
func convertLDAPError(err error) error {
	if err == nil {
		return nil
	}

	var ldapErr *ldap.Error
	if errors.As(err, &ldapErr) {
		switch ldapErr.ResultCode {
		case ldap.ErrorNetwork:
			// this one is especially important, because Teleport will
			// try to re-establish the connection when a ConnectionProblem
			// is detected
			return trace.ConnectionProblem(err, "network error")
		case ldap.LDAPResultOperationsError:
			if strings.Contains(err.Error(), "successful bind must be completed") {
				return trace.NewAggregate(trace.AccessDenied(
					"the LDAP server did not accept Teleport's client certificate, "+
						"has the Teleport CA been imported correctly?"), err)
			}
		case ldap.LDAPResultEntryAlreadyExists:
			return trace.AlreadyExists("LDAP object already exists: %v", err)
		case ldap.LDAPResultConstraintViolation:
			return trace.BadParameter("object constraint violation: %v", err)
		case ldap.LDAPResultInsufficientAccessRights:
			return trace.AccessDenied("insufficient permissions: %v", err)
		}
	}

	return err
}

// ReadWithFilter searches the specified DN (and its children) using the specified LDAP filter.
// See https://ldap.com/ldap-filters/ for more information on LDAP filter syntax.
func (c *LDAPClient) ReadWithFilter(dn string, filter string, attrs []string) ([]*ldap.Entry, error) {
	req := ldap.NewSearchRequest(
		dn,
		ldap.ScopeWholeSubtree,
		ldap.DerefAlways,
		0,     // no SizeLimit
		0,     // no TimeLimit
		false, // TypesOnly == false, we want attribute values
		filter,
		attrs,
		nil, // no Controls
	)
	c.mu.Lock()
	defer c.mu.Unlock()

	res, err := c.client.SearchWithPaging(req, searchPageSize)
	if err != nil {
		return nil, trace.Wrap(convertLDAPError(err), "fetching LDAP object %q with filter %q", dn, filter)
	}

	return res.Entries, nil
}

// Read fetches an LDAP entry at path and its children, if any. Only
// entries with the given class are returned and only with the specified
// attributes.
//
// You can browse LDAP on the Windows host to find the objectClass for a
// specific entry using ADSIEdit.msc.
// You can find the list of all AD classes at
// https://docs.microsoft.com/en-us/windows/win32/adschema/classes-all
func (c *LDAPClient) Read(dn string, class string, attrs []string) ([]*ldap.Entry, error) {
	return c.ReadWithFilter(dn, fmt.Sprintf("(%s=%s)", AttrObjectClass, class), attrs)
}

// Create creates an LDAP entry at the given path, with the given class and
// attributes. Note that AD will create a bunch of attributes for each object
// class automatically and you don't need to specify all of them.
//
// You can browse LDAP on the Windows host to find the objectClass and
// attributes for similar entries using ADSIEdit.msc.
// You can find the list of all AD classes at
// https://docs.microsoft.com/en-us/windows/win32/adschema/classes-all
func (c *LDAPClient) Create(dn string, class string, attrs map[string][]string) error {
	req := ldap.NewAddRequest(dn, nil)
	for k, v := range attrs {
		req.Attribute(k, v)
	}
	req.Attribute("objectClass", []string{class})

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.client.Add(req); err != nil {
		return trace.Wrap(convertLDAPError(err), "error creating LDAP object %q", dn)
	}
	return nil
}

// CreateContainer creates an LDAP container entry if
// it doesn't already exist.
func (c *LDAPClient) CreateContainer(dn string) error {
	err := c.Create(dn, classContainer, nil)
	// Ignore the error if container already exists.
	if trace.IsAlreadyExists(err) {
		return nil
	}

	return trace.Wrap(err)
}

// Update updates an LDAP entry at the given path, replacing the provided
// attributes. For each attribute in replaceAttrs, the value is completely
// replaced, not merged. If you want to modify the value of an existing
// attribute, you should read the existing value first, modify it and provide
// the final combined value in replaceAttrs.
//
// You can browse LDAP on the Windows host to find attributes of existing
// entries using ADSIEdit.msc.
func (c *LDAPClient) Update(dn string, replaceAttrs map[string][]string) error {
	req := ldap.NewModifyRequest(dn, nil)
	for k, v := range replaceAttrs {
		req.Replace(k, v)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.client.Modify(req); err != nil {
		return trace.Wrap(convertLDAPError(err), "updating %q", dn)
	}
	return nil
}

// CombineLDAPFilters joins the slice of filters
func CombineLDAPFilters(filters []string) string {
	return "(&" + strings.Join(filters, "") + ")"
}

func crlContainerDN(domain string, caType types.CertAuthType) string {
	return fmt.Sprintf("CN=%s,CN=CDP,CN=Public Key Services,CN=Services,CN=Configuration,%s", crlKeyName(caType), DomainDN(domain))
}

// CRLDN computes the distinguished name for a Teleport issuer in Windows environments.
func CRLDN(issuerID string, activeDirectoryDomain string, caType types.CertAuthType) string {
	return "CN=" + issuerID + "," + crlContainerDN(activeDirectoryDomain, caType)
}

// CRLDistributionPoint computes the CRL distribution point for certs issued.
func CRLDistributionPoint(activeDirectoryDomain string, caType types.CertAuthType, issuer *tlsca.CertAuthority, includeSKID bool) string {
	name := issuer.Cert.Subject.CommonName
	if includeSKID {
		id := base32.HexEncoding.EncodeToString(issuer.Cert.SubjectKeyId)
		name = id + "_" + name
	}
	crlDN := CRLDN(name, activeDirectoryDomain, caType)
	return fmt.Sprintf("ldap:///%s?certificateRevocationList?base?objectClass=cRLDistributionPoint", crlDN)
}

// crlKeyName returns the appropriate LDAP key given the CA type.
//
// Note: UserCA must use "Teleport" to keep backwards compatibility.
func crlKeyName(caType types.CertAuthType) string {
	switch caType {
	case types.DatabaseClientCA, types.DatabaseCA:
		return "TeleportDB"
	default:
		return "Teleport"
	}
}
