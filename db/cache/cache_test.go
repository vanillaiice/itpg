package cache

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

var DB *Cache

var dbUrl string

var testValues = map[string]string{
	"foo": "bar",
	"baz": "qux",
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatal(err)
	}

	if err = pool.Client.Ping(); err != nil {
		log.Fatal(err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "7.2.5-alpine",
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatal(err)
	}

	addr := net.JoinHostPort("localhost", resource.GetPort("6379/tcp"))
	dbUrl = fmt.Sprintf("redis://%s", addr)

	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		DB, err = New(dbUrl, context.Background())
		return err
	}); err != nil {
		log.Fatal(err)
	}

	defer DB.Close()

	code := m.Run()

	if err = pool.Purge(resource); err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func TestNew(t *testing.T) {
	testDB, err := New(dbUrl, context.Background())
	if err != nil {
		t.Error(err)
	}
	if err = testDB.Close(); err != nil {
		t.Error(err)
	}
}

func TestSet(t *testing.T) {
	for k, v := range testValues {
		err := DB.Set(k, v, time.Minute)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGet(t *testing.T) {
	for k, v := range testValues {
		val, err := DB.Get(k)
		if err != nil {
			t.Error(err)
		}
		if val != v {
			t.Errorf("got %s, want %s", val, v)
		}
	}
}

func TestClose(t *testing.T) {
	err := DB.Close()
	if err != nil {
		t.Error(err)
	}
}
