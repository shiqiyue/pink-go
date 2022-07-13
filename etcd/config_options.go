package etcd

// Code generated by github.com/launchdarkly/go-options.  DO NOT EDIT.

import (
	"fmt"
	"github.com/busgo/pink-go/log"
	"time"
)

import "github.com/google/go-cmp/cmp"

type ApplyOptionFunc func(c *config) error

func (f ApplyOptionFunc) apply(c *config) error {
	return f(c)
}

func newConfig(options ...Option) (config, error) {
	var c config
	err := applyConfigOptions(&c, options...)
	return c, err
}

func applyConfigOptions(c *config, options ...Option) error {
	for _, o := range options {
		if err := o.apply(c); err != nil {
			return err
		}
	}
	return nil
}

type Option interface {
	apply(*config) error
}

type optionEndpointsImpl struct {
	o []string
}

func (o optionEndpointsImpl) apply(c *config) error {
	c.endpoints = o.o
	return nil
}

func (o optionEndpointsImpl) Equal(v optionEndpointsImpl) bool {
	switch {
	case !cmp.Equal(o.o, v.o):
		return false
	}
	return true
}

func (o optionEndpointsImpl) String() string {
	name := "OptionEndpoints"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

func OptionEndpoints(o []string) Option {
	return optionEndpointsImpl{
		o: o,
	}
}

type optionUserNameImpl struct {
	o string
}

func (o optionUserNameImpl) apply(c *config) error {
	c.userName = o.o
	return nil
}

func (o optionUserNameImpl) Equal(v optionUserNameImpl) bool {
	switch {
	case !cmp.Equal(o.o, v.o):
		return false
	}
	return true
}

func (o optionUserNameImpl) String() string {
	name := "OptionUserName"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

func OptionUserName(o string) Option {
	return optionUserNameImpl{
		o: o,
	}
}

type optionPasswordImpl struct {
	o string
}

func (o optionPasswordImpl) apply(c *config) error {
	c.password = o.o
	return nil
}

func (o optionPasswordImpl) Equal(v optionPasswordImpl) bool {
	switch {
	case !cmp.Equal(o.o, v.o):
		return false
	}
	return true
}

func (o optionPasswordImpl) String() string {
	name := "OptionPassword"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

func OptionPassword(o string) Option {
	return optionPasswordImpl{
		o: o,
	}
}

type optionDialTimeoutImpl struct {
	o time.Duration
}

func (o optionDialTimeoutImpl) apply(c *config) error {
	c.dialTimeout = o.o
	return nil
}

func (o optionDialTimeoutImpl) Equal(v optionDialTimeoutImpl) bool {
	switch {
	case !cmp.Equal(o.o, v.o):
		return false
	}
	return true
}

func (o optionDialTimeoutImpl) String() string {
	name := "OptionDialTimeout"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

func OptionDialTimeout(o time.Duration) Option {
	return optionDialTimeoutImpl{
		o: o,
	}
}

type optionLImpl struct {
	o log.Logger
}

func (o optionLImpl) apply(c *config) error {
	c.l = o.o
	return nil
}

func (o optionLImpl) Equal(v optionLImpl) bool {
	switch {
	case !cmp.Equal(o.o, v.o):
		return false
	}
	return true
}

func (o optionLImpl) String() string {
	name := "OptionL"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

func OptionL(o log.Logger) Option {
	return optionLImpl{
		o: o,
	}
}
