package sshconf_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/lfaoro/ssm/pkg/sshconf"
)

func TestParse(t *testing.T) {
	cfg, err := sshconf.ParsePath("./config_test")
	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	for _, h := range cfg.Hosts {
		fmt.Println(h)
		fmt.Println(h.Name)
		for k, v := range h.Options {
			fmt.Println(k, v)
		}
	}
	_, err = sshconf.ParsePath("./nonexistent")
	if err == nil {
		t.FailNow()
	}
}
