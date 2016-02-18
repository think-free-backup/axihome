package variablemanager

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/boltdb/bolt"
	"github.com/think-free/log"
)

type VariableManager struct {
	db   *bolt.DB
	file string
}

func New(file string) *VariableManager {

	vm := &VariableManager{}

	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	vm.db = db
	vm.file = file

	runtime.SetFinalizer(vm, Close)

	return vm
}

func Close(vm *VariableManager) {

	log.Println("Closing " + vm.file)
	vm.db.Close()
}

func (vm *VariableManager) GetDB() *bolt.DB {

	return vm.db
}

func (vm *VariableManager) CreateBucket(bucket string) {

	vm.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			log.Println("Can't create", bucket, "bucket:", err)
		}
		return err
	})
}

func (vm *VariableManager) Set(bucket, name string, value interface{}) error {

	err := vm.db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(bucket))
		val, _ := json.Marshal(value)
		err := b.Put([]byte(name), val)
		if err != nil {
			log.Println("Can't write to "+bucket+" bucket:", err)
		}

		return err
	})

	return err
}

func (vm *VariableManager) SetJson(bucket, name, value string) error {

	err := vm.db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(name), []byte(value))
		if err != nil {
			log.Println("Can't write to "+bucket+" bucket:", err)
		}

		return err
	})

	return err
}

func (vm *VariableManager) SetAll(bucket string, content []byte) error {

	err := vm.db.Batch(func(tx *bolt.Tx) error {

		tx.DeleteBucket([]byte(bucket))
		b, err := tx.CreateBucket([]byte(bucket))

		var js interface{}
		json.Unmarshal(content, &js)

		conf := js.(map[string]interface{})

		for k, v := range conf {

			js, _ := json.Marshal(v)
			fmt.Println("\x1b[32m" + k + "\x1b[0m : " + string(js))
			fmt.Println("")
			b.Put([]byte(k), js)
		}

		return err
	})

	return err
}

func (vm *VariableManager) Get(bucket, name string) interface{} {

	var value interface{}

	vm.db.View(func(tx *bolt.Tx) error {

		bv := tx.Bucket([]byte(bucket))

		jsonvalue := bv.Get([]byte(name))
		json.Unmarshal(jsonvalue, &value)

		return nil
	})

	return value
}

func (vm *VariableManager) GetJson(bucket, name string) []byte {

	var jsonvalue []byte

	vm.db.View(func(tx *bolt.Tx) error {

		bv := tx.Bucket([]byte(bucket))

		jsonvalue = bv.Get([]byte(name))

		return nil
	})

	return jsonvalue
}

func (vm *VariableManager) GetAll(bucket string) map[string]interface{} {

	retmap := make(map[string]interface{})

	vm.db.View(func(tx *bolt.Tx) error {

		bv := tx.Bucket([]byte(bucket))

		if bv == nil {

			log.Println("Bucket not found")
			return nil
		}

		bv.ForEach(func(k, jsonvalue []byte) error {

			var value interface{}
			json.Unmarshal(jsonvalue, &value)
			retmap[string(k)] = value
			return nil
		})

		return nil
	})

	return retmap
}
