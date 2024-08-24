package model

import (
	"time"
)

type TLSConnection struct {
	Id              string    `json:"id"`
	Src             string    `json:"src"`
	SrcName         string    `json:"srcName"`
	SrcNamespace    string    `json:"srcNamespace"`
	Dst             string    `json:"dst"`
	DstName         string    `json:"dstName"`
	DstPort         uint16    `json:"dstPort"`
	Domain          string    `json:"domain"`
	UsedTLSVersion  string    `json:"usedTLSVersion"`
	UsedCipherSuite string    `json:"usedCipherSuite"`
	LastSeen        time.Time `json:"lastSeen"`
}

type Certificate struct {
	NotBefore   time.Time `json:"notBefore"`
	NotAfter    time.Time `json:"notAfter"`
	ServerChain string    `json:"serverChain"`
	LastScrape  time.Time `json:"lastScrape"`
}

type TLSDetails struct {
	Id                 string      `json:"id"`
	Domain             string      `json:"domain"`
	Dst                string      `json:"dst"`
	Port               uint16      `json:"port"`
	ClientTLSVersions  []string    `json:"clientTLSVersions"`
	ClientCipherSuites []string    `json:"clientCipherSuites"`
	UsedTLSVersion     string      `json:"usedTLSVersion"`
	UsedCipherSuite    string      `json:"usedCipherSuite"`
	Certificate        Certificate `json:"certificate"`
}
