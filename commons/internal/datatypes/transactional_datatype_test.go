package datatypes

import (
	"github.com/knowhunger/ortoo/commons/log"
	"sync"
	"testing"
	"time"
)

type AAA interface {
	AAA()
	AAATrans()
}

type AAAImpl struct {
}

func (a *AAAImpl) AAA() {
	log.Logger.Info("AAA")
}

func (a *AAAImpl) AAATrans() {
	log.Logger.Info("trans AAA")
}

type BBB interface {
	BBB()
	DoTransaction(trans func(bb BBB))
}

type BBBImpl struct {
	*AAAImpl
	trans *BBBImpl
}

func NewBBB() BBB {
	return &BBBImpl{
		AAAImpl: &AAAImpl{},
		trans:   nil,
	}
}

func (b *BBBImpl) BBB() {
	if b.trans != nil {
		b.AAATrans()
	} else {
		b.AAA()
	}
}

func (b *BBBImpl) DoTransaction(trans func(bb BBB)) {
	bb := &BBBImpl{
		AAAImpl: b.AAAImpl,
		trans:   b,
	}
	trans(bb)
}

func TestTTT(t *testing.T) {
	b := NewBBB()

	b.DoTransaction(func(bb BBB) {
		bb.BBB()
	})
	b.BBB()
}

func TestLockWorkingTest(t *testing.T) {
	var m = new(sync.RWMutex)
	var wg = new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		log.Logger.Println("begin func1")
		m.Lock()
		defer m.Unlock()
		defer wg.Done()
		log.Logger.Println("do func1")
		time.Sleep(3 * time.Second)
		log.Logger.Println("end func1")
	}()

	go func() {
		log.Logger.Println("begin func2")
		m.Lock()
		defer m.Unlock()
		defer wg.Done()
		log.Logger.Println("do func2")
		time.Sleep(2 * time.Second)
		log.Logger.Println("end func2")
	}()

	wg.Wait()
}
