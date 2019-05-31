package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/perlin-network/life/exec"
	"github.com/perlin-network/life/platform"
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
	wasmBytes, err := ioutil.ReadFile("transfer_back.wasm")
	if err != nil {
		panic(err)
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

	m, err := exec.NewModule(wasmBytes, exec.VMConfig{}, &importer{input: reasonBytes}, nil)
	if err != nil { // if the wasm bytecode is invalid
		panic(err)
	}

	entryID, ok := m.GetFunctionExport("contract_main") // can be changed to your own exported function
	if !ok {
		panic("entry function not found")
	}

	aotSvc := platform.FullAOTCompileModule(m)

	vm := m.NewVirtualMachine()
	vm.SetAOTService(aotSvc)

	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := vm.Run(entryID)
		if err != nil {
			vm.PrintStackTrace()
			panic(err)
		}
	}
	dur := time.Now().Sub(start)
	fmt.Printf("duration = %s\n", dur)
}

type importer struct {
	input []byte
}

func (i *importer) ResolveFunc(module, field string) exec.FunctionImport {
	//log.Printf("ResolveFunc %s:%s", module, field)
	switch module {
	case "env":
		switch field {
		case "_reason_len":
			return func(vm *exec.VirtualMachine) int64 {
				return int64(len(i.input))
			}
		case "_reason":
			return func(vm *exec.VirtualMachine) int64 {
				ptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				copy(vm.Memory[ptr:], i.input)
				return 0
			}
		case "_send_transaction":
			return func(vm *exec.VirtualMachine) int64 {
				tagOff := int(uint32(vm.GetCurrentFrame().Locals[0]))
				tagLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				payloadOff := int(uint32(vm.GetCurrentFrame().Locals[2]))
				payloadLen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				tag := string(vm.Memory[tagOff : tagOff+tagLen])
				payload := string(vm.Memory[payloadOff : payloadOff+payloadLen])
				log.Printf("tag = %q; payload = %q", tag, payload)
				return 0
			}
		case "_error":
			return func(vm *exec.VirtualMachine) int64 {
				offset := int(uint32(vm.GetCurrentFrame().Locals[0]))
				length := int(uint32(vm.GetCurrentFrame().Locals[1]))
				msg := string(vm.Memory[offset : offset+length])
				log.Printf("Error: " + msg)
				return 0
			}

		default:
			panic(fmt.Errorf("unknown import resolved: %s", field))
		}
	default:
		panic(fmt.Errorf("unknown module: %s", module))
	}
}

func (i *importer) ResolveGlobal(module, field string) int64 {
	log.Printf("ResolveGlobal %s:%s", module, field)
	return 0
}
