package config

import "time"

// config sürecinde çalıştırılması gereken funcs toparlandığı base func
func Run() {
	go func() {
		ticker := time.NewTicker(2* time.Hour)
		InitConfig()
		log.Fatal(err)
	} ()
}
