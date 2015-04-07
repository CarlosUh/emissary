// Password- or key-based SFTP delivery module
//
// PrivateKey must be already in PEM format

package sftp

import (
	"errors"
	gosftp "github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"path/filepath"
	"strings"
)

const (
	AUTH_MODE_PASSWORD = iota
	AUTH_MODE_KEY      = iota
)

type PathFunc func() string

type SFTP struct {
	AuthMode   int
	Username   string
	PrivateKey []byte
	Password   string
	Host       string
	// If path is provided, then it will be used. Otherwise it will delegate to PathFunc
	Path     string
	PathFunc PathFunc
}

func (s *SFTP) Deliver(r io.Reader) error {
	client, err := s.getSSHClient()

	if err != nil {
		return err
	}

	sftpclient, err := gosftp.NewClient(client)
	if err != nil {
		return err
	}

	var path string
	// Determine the path
	if len(s.Path) > 0 {
		path = s.Path
	} else if s.PathFunc != nil {
		path = s.PathFunc()
	} else {
		return errors.New("Missing path and path func")
	}

	// Create the path if necessary...
	if strings.Contains(path, "/") {
		dir := filepath.Dir(path)
		spl := strings.Split(dir, "/")
		currentFolder := ""
		for _, folder := range spl {
			if len(currentFolder) == 0 {
				currentFolder = folder
			} else {
				currentFolder = currentFolder + "/" + folder
			}

			info, err := sftpclient.Lstat(currentFolder)

			if err != nil {
				if strings.Contains(err.Error(), "SSH_FX_NO_SUCH_FILE") {
					err = sftpclient.Mkdir(currentFolder)
					if err != nil {
						return err
					}
				} else {
					return err
				}

			} else {
				if !info.IsDir() {
					return errors.New("Parent path exists but it is not a directory!")
				}
			}
		}
	}

	file, err := sftpclient.Create(path)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, r)

	return err
}

func (s *SFTP) getSSHClient() (*ssh.Client, error) {
	conf := &ssh.ClientConfig{
		User: s.Username,
	}
	switch s.AuthMode {
	case AUTH_MODE_KEY:
		if len(s.PrivateKey) == 0 {
			return &ssh.Client{}, errors.New("Missing private key for key-based authentication")
		}

		key, err := ssh.ParsePrivateKey(s.PrivateKey)
		if err != nil {
			return &ssh.Client{}, err
		}

		conf.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(key),
		}
	case AUTH_MODE_PASSWORD:
		if len(s.Password) == 0 {
			return &ssh.Client{}, errors.New("Password is required for password-based authentication")
		}
		conf.Auth = []ssh.AuthMethod{
			ssh.Password(s.Password),
		}
	default:
		return &ssh.Client{}, errors.New("Unknown auth mode")
	}

	return ssh.Dial("tcp", s.Host, conf)
}
