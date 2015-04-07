// Delivery module for plain FTP. Username/password is expected, and it probably
// will not work if it is anonymous/no authentication.
//
// Obligatory disclaimer: don't use this unless you're encrypting the data.

package ftp

import (
	"errors"
	jftp "github.com/jlaffaye/ftp"
	"io"
	"path/filepath"
	"strings"
	"time"
)

type PathFunc func() string

type FTP struct {
	Username string
	Password string

	// Address, excluding protocol, including port
	Address string

	// Timeout, in seconds
	Timeout int

	// If path is provided, then it will be used. Otherwise it will delegate to PathFunc
	Path     string
	PathFunc PathFunc
}

func (f *FTP) Deliver(r io.Reader) error {
	if f.Timeout == 0 {
		// Default to 5 second timeout
		f.Timeout = 5
	}

	conn, err := jftp.DialTimeout(f.Address, time.Duration(f.Timeout)*time.Second)

	if err != nil {
		return err
	}

	// Login
	err = conn.Login(f.Username, f.Password)
	if err != nil {
		return err
	}

	var path string
	// Determine the path
	if len(f.Path) > 0 {
		path = f.Path
	} else if f.PathFunc != nil {
		path = f.PathFunc()
	} else {
		return errors.New("Missing path and path func")
	}

	// Create the path if necessary...
	if strings.Contains(path, "/") {
		dir := filepath.Dir(path)
		dir = dir[1:]
		spl := strings.Split(dir, "/")

		for _, folder := range spl {
			files, err := conn.NameList(folder)
			if err != nil {
				return err
			}

			if len(files) == 0 {
				// Folder does not exist. Create
				err = conn.MakeDir(folder)
				if err != nil {
					return err
				}
			}

			err = conn.ChangeDir(folder)
			if err != nil {
				return err
			}
		}

		path = filepath.Base(path)
	}
	return conn.Stor(path, r)

}
