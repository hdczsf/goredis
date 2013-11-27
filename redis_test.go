package redis

import (
	"fmt"
	"testing"
	"time"
)

var (
	network = "tcp"
	address = "192.168.84.250:6379"
)

func dial() (*Redis, error) {
	return DialTimeout(network, address, 0, "", 5*time.Second, 5)
}

func TestDial(t *testing.T) {
	_, err := dial()
	if err != nil {
		t.Error(err)
	}
}

func TestDialFail(t *testing.T) {
	_, err := DialTimeout(network, address+"0", 0, "", 5*time.Second, 5)
	if err == nil {
		t.Error(err)
	}
}

func TestDiaURL(t *testing.T) {
	rawurl := fmt.Sprintf("redis://%s/1?size=5&timeout=10s", address)
	r, err := DialURL(rawurl)
	if err != nil {
		t.Fatal(err)
	}
	if r.db != 1 || r.size != 5 || r.timeout != 10*time.Second {
		t.Fail()
	}
}

func TestDialURLFail(t *testing.T) {
	rawurl := fmt.Sprintf("redis://tester:password@%s/1", address)
	_, err := DialURL(rawurl)
	if err == nil {
		t.Fail()
	}
}

func TestAuth(t *testing.T) {
	r, _ := dial()
	if err := r.Auth("password"); err == nil {
		t.Fail()
	}
}

func TestClientList(t *testing.T) {
	r, _ := dial()
	_, err := r.ClientList()
	if err != nil {
		t.Error(err)
	}
}

func TestAppend(t *testing.T) {
	r, _ := dial()
	r.Del("key")
	n, err := r.Append("key", "value")
	if err != nil {
		t.Error(err)
	}
	if n != 5 {
		t.Fail()
	}
	n, err = r.Append("key", "value")
	if err != nil {
		t.Error(err)
	}
	if n != 10 {
		t.Fail()
	}
	r.Del("key")
	r.LPush("key", "value")
	if _, err := r.Append("key", "value"); err == nil {
		t.Error(err)
	}
}

func TestBLPop(t *testing.T) {
	r, _ := dial()
	r.Del("key")
	result, err := r.BLPop([]string{"key"}, 1)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 0 {
		t.Fail()
	}
	r.LPush("key", "value")
	result, err = r.BLPop([]string{"key"}, 0)
	if err != nil {
		t.Error(err)
	}
	if len(result) == 0 {
		t.Fail()
	}
	if result[0] != "key" || result[1] != "value" {
		t.Fail()
	}
}

func TestDBSize(t *testing.T) {
	r, _ := dial()
	r.FlushDB()
	n, err := r.DBSize()
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Fail()
	}
}

func TestEval(t *testing.T) {
	r, _ := dial()
	rp, err := r.Eval("return {KEYS[1], KEYS[2], ARGV[1], ARGV[2]}", []string{"key1", "key2"}, []string{"arg1", "arg2"})
	if err != nil {
		t.Error(err)
	}
	l, err := r.listReturnValue(rp)
	if err != nil {
		t.Error(err)
	}
	if l[0] != "key1" || l[3] != "arg2" {
		t.Fail()
	}
	rp, err = r.Eval("return redis.call('set','foo','bar')", nil, nil)
	if err != nil {
		t.Error(err)
	}
	if err := r.okStatusReturnValue(rp); err != nil {
		t.Error(err)
	}
	rp, err = r.Eval("return 10", nil, nil)
	if err != nil {
		t.Error(err)
	}
	n, err := r.integerReturnValue(rp)
	if err != nil {
		t.Error(err)
	}
	if n != 10 {
		t.Fail()
	}
	rp, err = r.Eval("return {1,2,{3,'Hello World!'}}", nil, nil)
	if err != nil {
		t.Error(err)
	}
	if len(rp.Multi) != 3 {
		t.Fail()
	}
	if rp.Multi[2].Multi[0].Integer != 3 {
		t.Fail()
	}
	if s, err := r.stringBulkReturnValue(rp.Multi[2].Multi[1]); err != nil || s != "Hello World!" {
		t.Fail()
	}
}
