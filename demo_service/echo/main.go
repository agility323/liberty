package main

import (
	sf "github.com/agility323/liberty/service_framework"
)

func main() {
	sf.ServiceType = Conf.ServiceType
	sf.Start()
}
