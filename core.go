package seutils

import (
	"encoding/json"
	"github.com/nclk/selenium"
	//"github.com/tebeka/selenium"
	"fmt"
	"errors"
	"strings"
	"time"
)

type SeleniumConfiguration struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Concurrency int `json:"concurrency"`
	ImplicitWaitTimeout int `json:"implicit-wait-timeout"`
	Capabilities selenium.Capabilities `json:"capabilities"`
}

func NewDriverFromJSON(se_string string) (selenium.WebDriver, error) {
	var se SeleniumConfiguration
	if err := json.Unmarshal([]byte(se_string), &se); err != nil {
		return nil, err
	} else {
		return NewDriver(se)
	}
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

	implicit_wait_timeout, err := time.ParseDuration(fmt.Sprintf(
		`%ds`, se.ImplicitWaitTimeout,
	))
	if err != nil {
		driver.Quit()
		return nil, err
	}

	err = driver.SetImplicitWaitTimeout(implicit_wait_timeout)
	if err != nil {
		driver.Quit()
		return nil, err
	}

	return driver, nil

}

func QuerySelector(
	driver interface{
		FindElement(string, string) (selenium.WebElement, error)
	},
	by string,
	selector string,
	el_chan chan selenium.WebElement,
	err_chan chan error,
) {
	el, err := driver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		err_chan <- err
		close(el_chan)
	} else {
		el_chan <- el
	}
}

func GetLocation(
	element selenium.WebElement,
	point_chan chan *selenium.Point,
	err_chan chan error,
) {
	point, err := element.Location()
	if err != nil {
		err_chan <- err
		close(point_chan)
	} else {
		point_chan <- point
	}
}

func CheckCSSProperty(
	label string,
	element selenium.WebElement,
	property string,
	value string,
	done_chan chan bool,
	err_chan chan error,
) {
	attr, err := element.CSSProperty(property)
	if err != nil {
		err_chan <- err
	} else if (attr != value) {
		err_chan <- errors.New(fmt.Sprintf(
			`%s: CSS property { "%s": "%s" } failed to match value "%s"`,
			label, property, value, attr,
		))
	}
	done_chan <- true
}

func CheckAttribute(
	label string,
	el selenium.WebElement,
	name string,
	value string,
	done_chan chan bool,
	err_chan chan error,
) {
	attr, err := el.GetAttribute(name)
	if err != nil {
		if !strings.Contains(attr, value) {
			err_chan <- errors.New(fmt.Sprintf(
				`%s: %s ("%s") failed ` +
				`to contain "%s"`,
				label, name, attr, value,
			))
		}
	}
	done_chan <- true
}

func GetAttribute(
	el selenium.WebElement,
	name string,
	attr_chan chan string,
	err_chan chan error,
) {
	attr, err := el.GetAttribute(name)
	if err != nil {
		err_chan <- err
		close(attr_chan)
	} else {
		attr_chan <- attr
	}
}

