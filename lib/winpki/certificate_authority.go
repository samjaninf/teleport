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
	"context"
	"crypto/tls"
	"encoding/base32"
	"log/slog"

	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/tlsca"
)

// NewCertificateStoreClient returns a new structure for modifying windows certificates in a Windows CA.
func NewCertificateStoreClient(cfg CertificateStoreConfig) *CertificateStoreClient {
	return &CertificateStoreClient{cfg: cfg}
}

// CertificateStoreClient implements access to a Windows Certificate Authority
type CertificateStoreClient struct {
	cfg CertificateStoreConfig
}

// CRLGenerator generates CRLs, which are required for certificate-based authentication on Windows.
// Teleport has its own locking concept that is used for revocation, so the CRLS generated here
// are always empty and exist only to satisfy the Windows requirements for CRL checking.
type CRLGenerator interface {
	// GenerateCertAuthorityCRL returns an empty CRL for a CA.
	GenerateCertAuthorityCRL(ctx context.Context, caType types.CertAuthType) ([]byte, error)
	// GetCertAuthorities returns a list of cert authorities
	GetCertAuthorities(ctx context.Context, caType types.CertAuthType, loadKeys bool) ([]types.CertAuthority, error)
}

// CertificateStoreConfig is a config structure for a Windows Certificate Authority
type CertificateStoreConfig struct {
	// AccessPoint is the Auth API client (with caching).
	AccessPoint CRLGenerator
	// Domain is the Active Directory domain where Teleport publishes its
	// Certificate Revocation List (CRL).
	Domain string
	// Logger is the logging sink for the service
	Logger *slog.Logger
	// ClusterName is the name of this Teleport cluster
	ClusterName string
	// LC is the LDAPConfig
	LC *LDAPConfig
}

// Update publishes an empty certificate revocation list to LDAP.
func (c *CertificateStoreClient) Update(ctx context.Context, tc *tls.Config) error {
	caType := types.UserCA

	// TODO(zmb3): check for the presence of Teleport's CA in the NTAuth store

	// To make the CA trusted, we need 3 things:
	// 1. put the CA cert into the Trusted Certification Authorities in Group Policy
	// 2. put the CA cert into NTAuth store in LDAP
	// 3. put the CRL of the CA into a dedicated LDAP entry
	//
	// #1 and #2 are done manually as part of the set-up process (see public docs).
	// Below we do #3.

	hasCRL := false
	certAuthorities, err := c.cfg.AccessPoint.GetCertAuthorities(ctx, caType, false)
	if err != nil {
		return trace.Wrap(err)
	}
	for _, ca := range certAuthorities {
		for _, keyPair := range ca.GetActiveKeys().TLS {
			if len(keyPair.CRL) == 0 {
				continue
			}
			hasCRL = true
			cert, err := tlsca.ParseCertificatePEM(keyPair.Cert)
			if err != nil {
				return trace.Wrap(err)
			}
			subjectID := base32.HexEncoding.EncodeToString(cert.SubjectKeyId)
			issuer := subjectID + "_" + c.cfg.ClusterName
			if err := c.updateCRL(ctx, issuer, keyPair.CRL, caType, tc); err != nil {
				return trace.Wrap(err)
			}
		}
	}

	// All authorities are missing CRL, let's fall back to legacy behavior
	// TODO(probakowski): DELETE IN v21.0.0
	if !hasCRL {
		crlDER, err := c.cfg.AccessPoint.GenerateCertAuthorityCRL(ctx, caType)
		if err != nil {
			return trace.Wrap(err, "generating CRL")
		}

		if err := c.updateCRL(ctx, c.cfg.ClusterName, crlDER, caType, tc); err != nil {
			return trace.Wrap(err, "updating CRL over LDAP")
		}
	}
	return nil
}

func (c *CertificateStoreClient) updateCRL(ctx context.Context, issuer string, crlDER []byte, caType types.CertAuthType, tc *tls.Config) error {
	// Publish the CRL for current cluster CA. For trusted clusters, their
	// respective windows_desktop_services will publish CRLs of their CAs so we
	// don't have to do it here.
	//
	// CRLs live under the CDP (CRL Distribution Point) LDAP container. There's
	// another nested container with the CA name, I think, and then multiple
	// separate CRL objects in that container.
	//
	// We name our parent container based on the CA type (for example, for User
	// CA, it is called "Teleport"), and the CRL object is named after the
	// Teleport cluster name. So, for instance, CRL for cluster "prod" and User
	// CA will be placed at:
	// ... > CDP > Teleport > prod
	containerDN := crlContainerDN(c.cfg.Domain, caType)
	crlDN := CRLDN(issuer, c.cfg.Domain, caType)

	ldapClient, err := DialLDAP(ctx, c.cfg.LC, tc)
	if err != nil {
		return trace.Wrap(err, "dialing LDAP server")
	}
	defer ldapClient.Close()

	// Create the parent container.
	if err := ldapClient.CreateContainer(ctx, containerDN); err != nil {
		return trace.Wrap(err, "creating CRL container")
	}

	// Create the CRL object itself.
	if err := ldapClient.Create(
		crlDN,
		"cRLDistributionPoint",
		map[string][]string{"certificateRevocationList": {string(crlDER)}},
	); err != nil {
		if !trace.IsAlreadyExists(err) {
			return trace.Wrap(err)
		}
		// CRL already exists, update it.
		if err := ldapClient.Update(
			ctx,
			crlDN,
			map[string][]string{"certificateRevocationList": {string(crlDER)}},
		); err != nil {
			return trace.Wrap(err)
		}
		c.cfg.Logger.InfoContext(ctx, "Updated CRL for Windows logins via LDAP")
	} else {
		c.cfg.Logger.InfoContext(ctx, "Added CRL for Windows logins via LDAP")
	}
	return nil
}
