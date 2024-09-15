package main

import (
	"context"
	"crypto/tls"
	"fmt"
	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func init() {
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("root"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", "192.168.101.120:10022", config)
	if err != nil {
		log.Fatal("Error", err)
	}

	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Error", err)
	}

	defer session.Close()

	clientScp, err := scp.NewClientBySSH(client)
	if err != nil {
		fmt.Println("Error creating new SSH session from existing connection", err)
	}

	// Open a file
	f, _ := os.Open("../../k8spacket")

	// Close client connection after the file has been copied
	defer client.Close()

	// Close the file after it has been copied
	defer f.Close()

	// Finally, copy the file over
	// Usage: CopyFromFile(context, file, remotePath, permission)

	// the context can be adjusted to provide time-outs or inherit from other contexts if this is embedded in a larger application.
	err = clientScp.CopyFromFile(context.Background(), *f, "/root/k8spacket", "0655")

	if err != nil {
		fmt.Println("Error while copying file ", err)
	}

	session.Output("systemctl start k8spacket.service")

}

func TestNodegraphHeathEndpoint(t *testing.T) {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get("http://192.168.101.120:16676/nodegraph/api/health")
	if err != nil {
		fmt.Println(err)
	}
	assert.Eventually(t, func() bool {
		return resp.StatusCode == http.StatusOK
	}, time.Second*10, time.Millisecond*1000)

}
