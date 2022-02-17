package main

import (
	"github.com/kubasiemion/merkledemo/httpservice"

	"github.com/san-lab/commongo/gohttpservice"
)

func main() {
	h := httpservice.NewHandler()
	gohttpservice.Startserver(h)
}
