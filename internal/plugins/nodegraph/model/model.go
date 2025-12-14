package model

import "time"

type ConnectionItem struct {
	Src            string    `json:"src"`
	SrcName        string    `json:"srcName"`
	SrcNamespace   string    `json:"srcNamespace"`
	Dst            string    `json:"dst"`
	DstName        string    `json:"dstName"`
	DstNamespace   string    `json:"dstNamespace"`
	ConnCount      int64     `json:"connCount"`
	ConnPersistent int64     `json:"connPersistent"`
	BytesSent      float64   `json:"bytesSent"`
	BytesReceived  float64   `json:"bytesReceived"`
	Duration       float64   `json:"duration"`
	MaxDuration    float64   `json:"maxDuration"`
	LastSeen       time.Time `json:"lastSeen"`
}

type ConnectionEndpoint struct {
	Ip             string
	Name           string
	Namespace      string
	ConnCount      int64
	ConnPersistent int64
	BytesSent      float64
	BytesReceived  float64
	Duration       float64
	MaxDuration    float64
}

type NodeGraph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type DisplayConfig struct {
	DisplayName string `json:"displayName"`
	Color       string `json:"color"`
}

type Config struct {
	Arc1          DisplayConfig `json:"arc__1"`
	Arc2          DisplayConfig `json:"arc__2"`
	MainStat      DisplayConfig `json:"mainStat"`
	SecondaryStat DisplayConfig `json:"secondaryStat"`
}

type Node struct {
	Id            string  `json:"id"`
	Title         string  `json:"title"`
	SubTitle      string  `json:"subTitle"`
	MainStat      string  `json:"mainStat"`
	SecondaryStat string  `json:"secondaryStat"`
	Arc1          float64 `json:"arc__1"`
	Arc2          float64 `json:"arc__2"`
	Arc3          float64 `json:"arc__3"`
}

type Edge struct {
	Id            string `json:"id"`
	Source        string `json:"source"`
	Target        string `json:"target"`
	MainStat      string `json:"mainStat"`
	SecondaryStat string `json:"secondaryStat"`
}
