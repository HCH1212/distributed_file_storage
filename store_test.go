package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathKey := CASPathTransformFunc(key)
	expectedFilename := "6804429f74181a63c50c3d81d733a12f14a353ff"
	expectedPathName := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if pathKey.Pathname != expectedPathName {
		t.Errorf("have %s, want %s", pathKey.Pathname, expectedPathName)
	}
	if pathKey.Filename != expectedFilename {
		t.Errorf("have %s, want %s", pathKey.Filename, expectedFilename)
	}
}

// TestStore 主测试函数
func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	for i := 0; i < 50; i++ {

		key := fmt.Sprintf("foo_%d", i)
		data := []byte("some jpg bytes")

		// 写入
		if _, err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		// 是否存在
		if ok := s.Has(key); !ok {
			t.Errorf("expected to have key %s", key)
		}

		// 读取
		_, r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}

		b, _ := ioutil.ReadAll(r)
		if bytes.Compare(data, b) != 0 {
			t.Errorf("have %s, want %s", b, data)
		}
		fmt.Println(string(b))

		// 删除
		if err = s.Delete(key); err != nil {
			t.Error(err)
		}

		// 是否存在
		if ok := s.Has(key); ok {
			t.Errorf("expected to NOT have key %s", key)
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

// teardown 清除
func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
