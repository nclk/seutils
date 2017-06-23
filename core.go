package seutils

import (
	"github.com/tebeka/selenium"
	"fmt"
)

type SeleniumConfiguration struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Concurrency int `json:"concurrency"`
	Capabilities selenium.Capabilities `json:"capabilities"`
}

func NewDriver(se SeleniumConfiguration) (selenium.WebDriver, error) {

	var sePortString string
	if len(se.Port) > 0 {
		sePortString = ":" + se.Port
	}

	server := fmt.Sprintf(
		"http://%s%s/wd/hub",
		se.Host,
		sePortString)

	driver, err := selenium.NewRemote(
		se.Capabilities,
		server)

	if err != nil {
		return nil, err
	} else {
		return driver, nil
	}

}

