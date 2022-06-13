package integration

import (
	"bytes"
	ctx "context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/log"
	"github.com/orda-io/orda/server/redis"
	"github.com/orda-io/orda/server/utils"
	"github.com/stretchr/testify/require"
)

var wg = new(sync.WaitGroup)

func (its *IntegrationTestSuite) TestLock() {

	its.Run("Can lock with redis", func() {
		ts1 := time.Now()
		wg.Add(2)
		go its.tryRedisLock(its.T(), "cli1", true) // 3 seconds
		go its.tryRedisLock(its.T(), "cli2", true) // 3 seconds

		wg.Wait()
		ts2 := time.Now().Sub(ts1) // more than 6 seconds
		its.ctx.L().Infof("%v", ts2)
		require.Equal(its.T(), int(ts2.Seconds()), 6)

		wg.Add(2)
		its.tryRedisLock(its.T(), "cli3", false) // 3 seconds
		its.tryRedisLock(its.T(), "cli4", false) // 5 seconds
		wg.Wait()
		ts3 := time.Now().Sub(ts1)
		require.Equal(its.T(), int(ts3.Seconds()), 14)
	})

	its.Run("Can lock locally", func() {
		ts1 := time.Now()
		wg.Add(2)
		go its.tryLocalLock(its.T(), "cli1", true)
		go its.tryLocalLock(its.T(), "cli2", true)

		wg.Wait()
		ts2 := time.Now().Sub(ts1)
		its.ctx.L().Infof("%v", ts2)
		require.Equal(its.T(), int(ts2.Seconds()), 6)

		wg.Add(2)
		its.tryLocalLock(its.T(), "cli3", false) // 3 seconds
		its.tryLocalLock(its.T(), "cli4", false) // 5 seconds
		wg.Wait()
		ts3 := time.Now().Sub(ts1)
		require.Equal(its.T(), int(ts3.Seconds()), 14)
	})

	its.Run("Can patch documents concurrently", func() {

		jsonFormat := "{ \"json\": \"{\\\"hello\\\": %d}\"}"
		wg.Add(2)
		json1 := fmt.Sprintf(jsonFormat, 1)
		json2 := fmt.Sprintf(jsonFormat, 2)
		its.ctx.L().Infof("%s", json1)
		go func() {
			its.patch(its.T(), json1)
			wg.Done()
		}()

		go func() {
			its.patch(its.T(), json2)
			wg.Done()
		}()
		wg.Wait()
	})
}

func (its *IntegrationTestSuite) patch(t *testing.T, json string) {
	reqBody := bytes.NewBufferString(json)
	addr := fmt.Sprintf("http://localhost:%d/api/v1/collections/%s/documents/%s", its.conf.RestfulPort, its.collectionName, its.getTestName())
	its.ctx.L().Infof("%v", addr)
	resp, err := http.Post(addr, "application/json", reqBody)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		str := string(respBody)
		println(str)
	}
	time.Sleep(1 * time.Second)
}

func (its *IntegrationTestSuite) tryLocalLock(t *testing.T, cliName string, unlock bool) {
	ordaCtx := context.NewOrdaContext(ctx.TODO(), "test", cliName)

	defer func() {
		wg.Done()
	}()
	lock := utils.GetLocalLock(ordaCtx, its.getTestName())
	if lock.TryLock() {
		log.Logger.Infof("locked by %v", cliName)
		if unlock {
			defer func() {
				log.Logger.Infof("released by %v", cliName)
				lock.Unlock()
			}()
		}
		time.Sleep(3 * time.Second)
	}

}

func (its *IntegrationTestSuite) tryRedisLock(t *testing.T, cliName string, unlock bool) {

	ordaCtx := context.NewOrdaContext(ctx.TODO(), "test", cliName)

	cli1, err1 := redis.New(ordaCtx, &its.conf.Redis)
	require.NoError(t, err1)
	defer func() {
		if err := cli1.Close(); err != nil {
			its.T().Fail()
		}
		wg.Done()
	}()

	lock := cli1.GetLock(its.ctx, its.getTestName())

	log.Logger.Infof("try to lock at %v", cliName)
	if lock.TryLock() {
		if unlock {
			defer func() {
				lock.Unlock()
				log.Logger.Infof("released by %v", cliName)
			}()
		}

		log.Logger.Infof("locked by %v", cliName)
		time.Sleep(3 * time.Second)
	}

}
