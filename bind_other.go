// +build !linux

package iping

import "errors"

func bindInterface(fd int, ifname string) error {
	return errors.New("Bind to interface is not supported by operating system")
}
