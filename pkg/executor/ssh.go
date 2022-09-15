package executor

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
	"os"
	"sync"
)

type SSHExecutor struct {
	//Host     string
	//Port     int
	User         string
	Password     string
	IdentityFile string
	clients      map[string]*ssh.Client

	mutex sync.RWMutex
}

func NewSSHExecutor(user, password, keyPath string) *SSHExecutor {
	//sshCfg, err := newSSHConfig(user, password, keyPath)
	//if err != nil {
	//	return nil, err
	//}

	//address := fmt.Sprintf("%s:%d", host, globalCfg.SshPort)
	//client, err := ssh.Dial("tcp", address, sshCfg)
	//if err != nil {
	//	return nil, err
	//}

	return &SSHExecutor{
		//Host:     host,
		//Port:     globalCfg.SshPort,
		User:         user,
		Password:     password,
		IdentityFile: keyPath,
		clients:      make(map[string]*ssh.Client),
	}
}

func (s *SSHExecutor) Execute(address, cmd string) (string, error) {
	client, err := s.getClient(address)
	if err != nil {
		return "", err
	}

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf
	if err = session.Run(cmd); err != nil {
		return stderrBuf.String(), err
	}
	return stdoutBuf.String(), nil
}

func (s *SSHExecutor) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for addr, client := range s.clients {
		if err := client.Close(); err != nil {
			fmt.Printf("close ssh client err, host: %s, err: %+v\n", addr, err)
		}
	}
	s.clients = make(map[string]*ssh.Client)
}

func (s *SSHExecutor) Transfer(address, localPath, remotePath string) error {
	client, err := s.getClient(address)
	if err != nil {
		return err
	}

	scpClient, err := scp.NewClientBySSH(client)
	if err != nil {
		return err
	}

	// Open a file
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer scpClient.Session.Close()
	defer f.Close()

	if err = scpClient.CopyFromFile(context.Background(), *f, remotePath, "0755"); err != nil {
		return err
	}

	return nil
}

func (s *SSHExecutor) Download(localPath, remotePath string) error {
	//client, err := scp.NewClientBySSH(s.client)
	//if err != nil {
	//	return err
	//}
	//
	//// Open a file
	//f, err := os.Open(localPath)
	//if err != nil {
	//	return err
	//}
	//defer client.Session.Close()
	//defer f.Close()
	//
	//if err = client.CopyFromRemote(context.Background(), f, remotePath); err != nil {
	//	return err
	//}

	return nil
}

func (s *SSHExecutor) getClient(address string) (*ssh.Client, error) {
	s.mutex.RLock()
	if client, ok := s.clients[address]; ok {
		s.mutex.RUnlock()
		return client, nil
	}
	s.mutex.RUnlock()

	client, err := s.newClient(address)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (s *SSHExecutor) newClient(address string) (*ssh.Client, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sshCfg, err := newSSHConfig(s.User, s.Password, s.IdentityFile)
	if err != nil {
		return nil, fmt.Errorf("generate ssh config for %s error: %s", address, err.Error())
	}
	client, err := ssh.Dial("tcp", address, sshCfg)
	if err != nil {
		return nil, fmt.Errorf("create ssh client for address %s error: %s", address, err.Error())
	}
	fmt.Printf("create ssh client to %s\n", address)
	s.clients[address] = client
	return client, nil
}

func newSSHConfig(user, password, keyPath string) (*ssh.ClientConfig, error) {
	var auth []ssh.AuthMethod

	if len(password) > 0 {
		auth = append(auth, ssh.Password(password))
	}

	if len(keyPath) > 0 {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, err
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}

		auth = append(auth, ssh.PublicKeys(signer))
	}

	return &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}
