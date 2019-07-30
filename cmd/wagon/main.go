package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"time"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
)

type Reason struct {
	Kind    string      `json:"kind"`
	Sender  []int8      `json:"sender,omitempty"`
	Details interface{} `json:"details"`
}

type TransferActivation struct {
	Sender []int8 `json:"sender,omitempty"`
	Amount uint64 `json:"amount"`
}

func main() {
	raw, err := ioutil.ReadFile("function.wasm")
	//raw, err := ioutil.ReadFile("transfer_back.wasm")
	if err != nil {
		log.Fatalf("could not compile wast file: %v", err)
	}

	reason := Reason{
		Kind:   "transfer",
		Sender: []int8{0, 1, 2, 3},
		Details: TransferActivation{
			Sender: []int8{0, 1, 2, 3},
			Amount: 100000,
		},
	}
	reasonBytes, err := json.Marshal(&reason)
	if err != nil {
		log.Fatalf("could not marshal json: %v", err)
	}
	log.Println(string(reasonBytes))

	m, err := wasm.ReadModule(bytes.NewReader(raw), func(name string) (*wasm.Module, error) {
		// ReadModule takes as a second argument an optional "importer" function
		// that is supposed to locate and import other modules when some module is
		// requested (by name.)
		// Theoretically, a general "importer" function not unlike the Python's 'import'
		// mechanism (that tries to locate and import modules from a $PYTHONPATH)
		// could be devised.
		switch name {
		case "env":
			reasonLen := func(proc *exec.Process) int32 {
				//log.Println("reasonLen called")
				return int32(len(reasonBytes))
			}
			reason := func(proc *exec.Process, v int32) {
				//log.Println("reason called")
				proc.WriteAt(reasonBytes, int64(v))
			}
			sendTransaction := func(proc *exec.Process, tagOff, tagLen, payloadOff, payloadLen int32) {
				//log.Println("sendTransaction called")
				//log.Printf("args = %d, %d, %d, %d", tagOff, tagLen, payloadOff, payloadLen)

				tagBytes := make([]byte, tagLen)
				_, err := proc.ReadAt(tagBytes, int64(tagOff))
				if err != nil {
					log.Println(err)
					return
				}

				payloadBytes := make([]byte, payloadLen)
				_, err = proc.ReadAt(payloadBytes, int64(payloadOff))
				if err != nil {
					log.Println(err)
					return
				}
				log.Printf("tag = %q; payload = %q", string(tagBytes), string(payloadBytes))
			}
			errorFn := func(proc *exec.Process, ptr, len int32) {
				log.Println("error called")
				//log.Printf("args = %d, %d", ptr, len)
				p := make([]byte, len)
				_, err := proc.ReadAt(p, int64(ptr))
				if err != nil {
					log.Println(err)
					return
				}
				log.Println(string(p))
			}

			m := wasm.NewModule()
			m.Types = &wasm.SectionTypes{
				Entries: []wasm.FunctionSig{
					{
						Form:        0, // value for the 'func' type constructor
						ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
					},
					{
						Form:       0, // value for the 'func' type constructor
						ParamTypes: []wasm.ValueType{wasm.ValueTypeI32},
					},
					{
						Form:       0, // value for the 'func' type constructor
						ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32},
					},
					{
						Form:       0, // value for the 'func' type constructor
						ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32},
					},
				},
			}
			m.FunctionIndexSpace = []wasm.Function{
				{
					Sig:  &m.Types.Entries[0],
					Host: reflect.ValueOf(reasonLen),
					Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
				},
				{
					Sig:  &m.Types.Entries[1],
					Host: reflect.ValueOf(reason),
					Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
				},
				{
					Sig:  &m.Types.Entries[2],
					Host: reflect.ValueOf(sendTransaction),
					Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
				},
				{
					Sig:  &m.Types.Entries[3],
					Host: reflect.ValueOf(errorFn),
					Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
				},
			}
			m.Export = &wasm.SectionExports{
				Entries: map[string]wasm.ExportEntry{
					"_reason_len": {
						FieldStr: "_reason_len",
						Kind:     wasm.ExternalFunction,
						Index:    0,
					},
					"_reason": {
						FieldStr: "_reason",
						Kind:     wasm.ExternalFunction,
						Index:    1,
					},
					"_send_transaction": {
						FieldStr: "_send_transaction",
						Kind:     wasm.ExternalFunction,
						Index:    2,
					},
					"_error": {
						FieldStr: "_error",
						Kind:     wasm.ExternalFunction,
						Index:    3,
					},
				},
			}

			return m, nil
		}
		return nil, fmt.Errorf("module %q unknown", name)
	})
	if err != nil {
		log.Fatalf("could not read module: %v", err)
	}

	for name, entry := range m.Export.Entries {
		if entry.Kind == wasm.ExternalFunction {
			log.Printf("%s %s %d", name, entry.FieldStr, entry.Index)
		}
	}
	contractFunc, ok := m.Export.Entries["contract_main"]
	if !ok {
		log.Fatalln("could not find contract_main")
	}

	vm, err := exec.NewVM(m, exec.EnableAOT(true))
	if err != nil {
		log.Fatalf("could not create wagon vm: %v", err)
	}

	start := time.Now()
	for i := 0; i < 50000; i++ {
		_, err = vm.ExecCode(int64(contractFunc.Index))
		if err != nil {
			log.Fatalf("could not execute func(): %v", err)
		}
	}
	dur := time.Now().Sub(start)
	fmt.Printf("duration = %s\n", dur)
}
