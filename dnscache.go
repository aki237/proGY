package main

import (
	"net"

	"github.com/boltdb/bolt"
)

// Cache is a struct containing a boltdb object and the cachefile pointing to that file
type Cache struct {
	domainMap *bolt.DB // bolt.DB object
	cachefile string   // file pointing to that boltdb file
}

// NewCache returns a Cache type object and error, if there is any error in opening the cachefile
func NewCache(cachefile string) (Cache, error) {
	a := Cache{cachefile: cachefile}
	var err error
	a.domainMap, err = bolt.Open(cachefile, 0644, nil)
	return a, err
}

// LookupIP returns the IP string for a passed domain name and an error
// if there is any error in fetching data from the cachefile or lookup from the DNS.
// It is a method of Cache struct
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

			var ipv4 []byte
			for _, val := range ips {
				ip4 := val.To4()
				if ip4 != nil {
					ipv4 = []byte(ip4.String())
					break
				}
			}
			err = b.Put([]byte(domain), ipv4)
			IP = ips[0].String()
			return err
		}
		IP = string(ip)
		return nil
	})
	return IP, err
}
