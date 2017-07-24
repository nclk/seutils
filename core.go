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
		err_chan <- errors.New(fmt.Sprintf(
			`Finding "%s": %s`, selector, err.Error(),
		))
		close(el_chan)
	} else {
		el_chan <- el
	}
}

func QuerySelectorAll(
	driver interface{
		FindElements(string, string) ([]selenium.WebElement, error)
	},
	by string,
	selector string,
	el_chan chan selenium.WebElement,
	err_chan chan error,
) {
	els, err := driver.FindElements(selenium.ByCSSSelector, selector)
	if err != nil {
		err_chan <- errors.New(fmt.Sprintf(
			`Finding "%s": %s`, selector, err.Error(),
		))
		close(el_chan)
	} else {
		for i := 0; i < len(els); i++ {
			el_chan <- els[i]
		}
		close(el_chan)
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
	selector string,
	element selenium.WebElement,
	property string,
	value string,
	job *PromiseStream,
	err_chan chan error,
) {
	attr, err := element.CSSProperty(property)
	if err != nil {
		err_chan <- err
		job.Done(false)
	} else if (attr != value) {
		err_chan <- errors.New(fmt.Sprintf(
			`"%s": CSS property { "%s": "%s" } failed to match value "%s"`,
			selector, property, attr, value,
		))
		job.Done(false)
	} else {
		job.Done(true)
	}
}

func CheckAttribute(
	label string,
	el selenium.WebElement,
	name string,
	value string,
	job *PromiseStream,
	err_chan chan error,
) {
	attr, err := el.GetAttribute(name)
	if err != nil {
		err_chan <- errors.New(fmt.Sprintf(
			`"%s": Failed to get attribute "%s": %s`,
			label, name, err.Error(),
		))
	} else if !strings.Contains(attr, value) {
		err_chan <- errors.New(fmt.Sprintf(
			`"%s": attribute %s ("%s") failed ` +
			`to equal "%s"`,
			label, name, attr, value,
		))
	}
	job.Done(true)
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

type PromiseStream struct {
	Chan chan interface{}
	Jobs int
}

func (worker *PromiseStream) New() *PromiseStream {
	worker.Jobs = worker.Jobs + 1
	return worker
}

func (worker *PromiseStream) Done(status interface{}) {
	worker.Chan <- status
}

func NewPromiseStream() *PromiseStream {
	return &PromiseStream{make(chan interface{}), 0}
}

func (worker *PromiseStream) Take(count int) (interface{}, bool) {
	ret := make([]interface{}, 0)
	for ; count > 0; count-- {
		candidate, ok := <-worker.Chan
		if !ok {
			return ret, ok
		}
		ret = append(ret, candidate)
		worker.Jobs--
		if worker.Jobs < 1 {
			break
		}
	}
	return ret, true
}

func (worker *PromiseStream) Close() (interface{}, bool) {
	final := make([]interface{}, 0)
	for ; worker.Jobs > 0; worker.Jobs-- {
		candidate, ok := <-worker.Chan
		if !ok {
			close(worker.Chan)
			return final, ok
		}
		final = append(final, candidate)
		if worker.Jobs < 1 {
			close(worker.Chan)
			break
		}
	}
	return final, true
}

