package etcdclient

import "errors"

const (
	keyEtcdProfile = "/common/"

	KeyEtcdLockProfile = keyEtcdProfile + "lock/"
	KeyEtcdLock        = KeyEtcdLockProfile + "%s"
)

var (
	ErrValueMayChanged = errors.New("The value has been changed by others on this time.")
	ErrEtcdNotInit     = errors.New("etcd is not initialized")
)
