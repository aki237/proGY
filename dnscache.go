package main

import (
	"net"

	"github.com/boltdb/bolt"
)

type Cache struct {
	domainMap *bolt.DB
	cachefile string
}

func NewCache(cachefile string) (Cache, error) {
	a := Cache{cachefile: cachefile}
	var err error
	a.domainMap, err = bolt.Open(cachefile, 0644, nil)
	return a, err
}

func (d *Cache) LookupIP(domain string) (string, error) {
	IP := ""
	err := d.domainMap.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("dnscache"))
		if err != nil {
			return err
		}
		ip := b.Get([]byte(domain))
		if ip == nil {
			ips, err := net.LookupIP(domain)
			if err != nil || len(ips) < 1 {
				return err
			}
			err = b.Put([]byte(domain), []byte(ips[0].String()))
			IP = ips[0].String()
			return err
		}
		IP = string(ip)
		return nil
	})
	return IP, err
}
