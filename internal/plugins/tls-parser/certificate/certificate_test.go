package certificate

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/internal/infra/network"
	"github.com/k8spacket/k8spacket/internal/plugins/tls-parser/model"
	"github.com/stretchr/testify/assert"
)

var certPem = `-----BEGIN CERTIFICATE-----
MIID9zCCAt+gAwIBAgIUGZbgCXci7yPvgzWU24qvkWW+0kYwDQYJKoZIhvcNAQEL
BQAwgYoxCzAJBgNVBAYTAlBMMQ0wCwYDVQQIDARXTEtQMQ8wDQYDVQQHDAZQb3pu
YW4xEjAQBgNVBAoMCWs4c3BhY2tldDEMMAoGA1UECwwDc2VjMRUwEwYDVQQDDAxr
OHNwYWNrZXQuaW8xIjAgBgkqhkiG9w0BCQEWE2s4c3BhY2tldEBnbWFpbC5jb20w
HhcNMjQwOTAyMTc1MTMzWhcNMjQxMDAyMTc1MTMzWjCBijELMAkGA1UEBhMCUEwx
DTALBgNVBAgMBFdMS1AxDzANBgNVBAcMBlBvem5hbjESMBAGA1UECgwJazhzcGFj
a2V0MQwwCgYDVQQLDANzZWMxFTATBgNVBAMMDGs4c3BhY2tldC5pbzEiMCAGCSqG
SIb3DQEJARYTazhzcGFja2V0QGdtYWlsLmNvbTCCASIwDQYJKoZIhvcNAQEBBQAD
ggEPADCCAQoCggEBAKWVWbbTMH5cKgFidt59aw/hJT8fK2ujKo/v0adyC6mupA2J
3C71+5nXCiO5vFvXEPWdocOvlo+uhQXEiIjE3FSbNHdRzuIE2ou69utoLMcyLsm6
LBCV0zTR+Z8m+0u/FElsz2W8cL/KrPNpZ9Zz9qnafAoJB/rbb27jlZhl78RwNQ/L
AWfDy84TwaYEzRFJk2NbNhgwvz681oftNGGmqxg9mmMrp2ivINu2YqYpM5Jlp2Zl
CSAnZSUqPnvT9NTPCLPOl2aoc/Ysae58NpHk9Zc2IBUnuzlv5Vk8R+arpurNekGq
Jq/W1dceZnoOauULpG3yKfKM9UrSx/LdUAqU2z0CAwEAAaNTMFEwHQYDVR0OBBYE
FKC8BPHMWDVs3AYkyWw8RYadzvigMB8GA1UdIwQYMBaAFKC8BPHMWDVs3AYkyWw8
RYadzvigMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBADtXz39I
KD+BRXQC7znEZvTnEZEkXkw3MRKpPJgpXteYbF0PSSj9Jacz4xrdPKb8nXtFI8+T
MXJLDH1iQAgex72Yv1EbdVsGdNpnLpJqyUpGIqPVCMZZdcPeIsL+YxvB/Srd2vvR
aTvUvLPthdHDqCn44xYq0eP3WXZv1kgkE49KH+6tl4axs9iJM6zoqYqL+wGgjZYO
FRxJdrIL9tIvtd3wK9sYTagqmzJDHknovfg1A5PbUdi07sSMmT6aZiMJYz3knk0Z
tVqCx005gbHTe+QOGvGSWbDdxp0MGoIHqKPf4mc0cuYH1fooAi9dca5DiXjMAvL8
mChyT2om1IB6MVE=
-----END CERTIFICATE-----`

var want = `Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number: 146089397423325118095355643034145370809477026374 (0x1996e0097722ef23ef833594db8aaf9165bed246)
    Signature Algorithm: SHA256-RSA
        Issuer: C=PL,ST=WLKP,UnknownOID=2.5.4.7,O=k8spacket,OU=sec,CN=k8spacket.io,emailAddress=k8spacket@gmail.com
        Validity
            Not Before: Sep 2 17:51:33 2024 UTC
            Not After : Oct 2 17:51:33 2024 UTC
        Subject: C=PL,ST=WLKP,UnknownOID=2.5.4.7,O=k8spacket,OU=sec,CN=k8spacket.io,emailAddress=k8spacket@gmail.com
        Subject Public Key Info:
            Public Key Algorithm: RSA
                Public-Key: (2048 bit)
                Modulus:
                    a5:95:59:b6:d3:30:7e:5c:2a:01:62:76:de:7d:6b:
                    0f:e1:25:3f:1f:2b:6b:a3:2a:8f:ef:d1:a7:72:0b:
                    a9:ae:a4:0d:89:dc:2e:f5:fb:99:d7:0a:23:b9:bc:
                    5b:d7:10:f5:9d:a1:c3:af:96:8f:ae:85:05:c4:88:
                    88:c4:dc:54:9b:34:77:51:ce:e2:04:da:8b:ba:f6:
                    eb:68:2c:c7:32:2e:c9:ba:2c:10:95:d3:34:d1:f9:
                    9f:26:fb:4b:bf:14:49:6c:cf:65:bc:70:bf:ca:ac:
                    f3:69:67:d6:73:f6:a9:da:7c:0a:09:07:fa:db:6f:
                    6e:e3:95:98:65:ef:c4:70:35:0f:cb:01:67:c3:cb:
                    ce:13:c1:a6:04:cd:11:49:93:63:5b:36:18:30:bf:
                    3e:bc:d6:87:ed:34:61:a6:ab:18:3d:9a:63:2b:a7:
                    68:af:20:db:b6:62:a6:29:33:92:65:a7:66:65:09:
                    20:27:65:25:2a:3e:7b:d3:f4:d4:cf:08:b3:ce:97:
                    66:a8:73:f6:2c:69:ee:7c:36:91:e4:f5:97:36:20:
                    15:27:bb:39:6f:e5:59:3c:47:e6:ab:a6:ea:cd:7a:
                    41:aa:26:af:d6:d5:d7:1e:66:7a:0e:6a:e5:0b:a4:
                    6d:f2:29:f2:8c:f5:4a:d2:c7:f2:dd:50:0a:94:db:
                    3d
                Exponent: 65537 (0x10001)
        X509v3 extensions:
            X509v3 Subject Key Identifier:
                A0:BC:04:F1:CC:58:35:6C:DC:06:24:C9:6C:3C:45:86:9D:CE:F8:A0
            X509v3 Authority Key Identifier:
                keyid:A0:BC:04:F1:CC:58:35:6C:DC:06:24:C9:6C:3C:45:86:9D:CE:F8:A0
            X509v3 Basic Constraints: critical
                CA:TRUE
    Signature Algorithm: SHA256-RSA
         3b:57:cf:7f:48:28:3f:81:45:74:02:ef:39:c4:66:f4:e7:11:
         91:24:5e:4c:37:31:12:a9:3c:98:29:5e:d7:98:6c:5d:0f:49:
         28:fd:25:a7:33:e3:1a:dd:3c:a6:fc:9d:7b:45:23:cf:93:31:
         72:4b:0c:7d:62:40:08:1e:c7:bd:98:bf:51:1b:75:5b:06:74:
         da:67:2e:92:6a:c9:4a:46:22:a3:d5:08:c6:59:75:c3:de:22:
         c2:fe:63:1b:c1:fd:2a:dd:da:fb:d1:69:3b:d4:bc:b3:ed:85:
         d1:c3:a8:29:f8:e3:16:2a:d1:e3:f7:59:76:6f:d6:48:24:13:
         8f:4a:1f:ee:ad:97:86:b1:b3:d8:89:33:ac:e8:a9:8a:8b:fb:
         01:a0:8d:96:0e:15:1c:49:76:b2:0b:f6:d2:2f:b5:dd:f0:2b:
         db:18:4d:a8:2a:9b:32:43:1e:49:e8:bd:f8:35:03:93:db:51:
         d8:b4:ee:c4:8c:99:3e:9a:66:23:09:63:3d:e4:9e:4d:19:b5:
         5a:82:c7:4d:39:81:b1:d3:7b:e4:0e:1a:f1:92:59:b0:dd:c6:
         9d:0c:1a:82:07:a8:a3:df:e2:67:34:72:e6:07:d5:fa:28:02:
         2f:5d:71:ae:43:89:78:cc:02:f2:fc:98:28:72:4f:6a:26:d4:
         80:7a:31:51`

type mockNetwork struct {
	network.INetwork
}

func (mockNetwork *mockNetwork) IsDomainReachable(domain string) bool {
	if domain == "unreachable" {
		return false
	}
	return true
}

func (mockNetwork *mockNetwork) GetPeerCertificates(address string, port uint16) ([]*x509.Certificate, error) {
	if port == 100 {
		return []*x509.Certificate{}, errors.New("wrong port")
	}
	if address == "unreachable" {
		return []*x509.Certificate{}, errors.New("unreachable")
	}

	block, errBytes := pem.Decode([]byte(certPem))
	if block == nil {
		fmt.Println(errBytes)
		return nil, errors.New(string(errBytes))
	}
	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return []*x509.Certificate{x509Cert}, nil
}

func TestUpdateCertificateInfo(t *testing.T) {

	os.Setenv("K8S_PACKET_TLS_CERTIFICATE_CACHE_TTL", "1h")

	var tests = []struct {
		scenario           string
		oldValue, newValue model.TLSDetails
		want               string
	}{
		{"empty old", model.TLSDetails{Certificate: model.Certificate{LastScrape: time.Now().Add(time.Hour * -2)}},
			model.TLSDetails{Domain: "k8spacket.io", Port: 443,
				Certificate: model.Certificate{NotBefore: time.Now().Add(time.Hour * -1), NotAfter: time.Now().Add(time.Hour * 1)}}, want},
		{"ttl", model.TLSDetails{Certificate: model.Certificate{LastScrape: time.Now().Add(time.Minute * -2), ServerChain: want}},
			model.TLSDetails{Domain: "k8spacket.io", Port: 443,
				Certificate: model.Certificate{NotBefore: time.Now().Add(time.Hour * -1), NotAfter: time.Now().Add(time.Hour * 1)}}, want},
		{"invalid port", model.TLSDetails{Certificate: model.Certificate{}},
			model.TLSDetails{Domain: "k8spacket.io", Port: 0,
				Certificate: model.Certificate{NotBefore: time.Now().Add(time.Hour * -1), NotAfter: time.Now().Add(time.Hour * 1)}}, "UNAVAILABLE"},
		{"unreachable domain", model.TLSDetails{Certificate: model.Certificate{}},
			model.TLSDetails{Domain: "unreachable", Dst: "reachable", Port: 443,
				Certificate: model.Certificate{NotBefore: time.Now().Add(time.Hour * -1), NotAfter: time.Now().Add(time.Hour * 1)}}, want},
		{"wrong port", model.TLSDetails{Certificate: model.Certificate{}},
			model.TLSDetails{Domain: "unreachable", Dst: "reachable", Port: 100,
				Certificate: model.Certificate{NotBefore: time.Now().Add(time.Hour * -1), NotAfter: time.Now().Add(time.Hour * 1)}}, want},
		{"unreachable domain and IP", model.TLSDetails{Certificate: model.Certificate{}},
			model.TLSDetails{Domain: "unreachable", Dst: "unreachable", Port: 100,
				Certificate: model.Certificate{NotBefore: time.Now().Add(time.Hour * -1), NotAfter: time.Now().Add(time.Hour * 1)}}, "UNAVAILABLE"},
	}

	mockNetwork := &mockNetwork{}

	certificate := &Certificate{mockNetwork}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			certificate.UpdateCertificateInfo(&test.newValue, &test.oldValue)

			assert.EqualValues(t, strings.TrimSpace(test.want), strings.TrimSpace(test.newValue.Certificate.ServerChain))
		})
	}

}
