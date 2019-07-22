package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"
	"unsafe"

	wasm "github.com/wasmerio/go-ext-wasm/wasmer"
)

// #include <stdlib.h>
//
// extern int32_t _reason_len(void *context);
// extern void _reason(void *context, int32_t ptr);
// extern void _send_transaction(void *context, int32_t tag_ptr, int32_t tag_len, int32_t payload_ptr, int32_t payload_len);
// extern void _error(void *context, int32_t ptr, int32_t len);
import "C"

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
	// Reads the WebAssembly module as bytes.
	bytes, err := wasm.ReadBytes("transfer_back.wasm")
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

	imports := wasm.NewImports()
	imports.Append("_reason_len", _reason_len, C._reason_len)
	imports.Append("_reason", _reason, C._reason)
	imports.Append("_send_transaction", _send_transaction, C._send_transaction)
	imports.Append("_error", _error, C._error)

	module, err := wasm.Compile(bytes)
	if err != nil {
		panic(err)
	}
	defer module.Close()

	// Instantiates the WebAssembly module.
	instance, err := module.InstantiateWithImports(imports)
	if err != nil {
		panic(err)
	}
	defer instance.Close()

	hdr := *(*reflect.SliceHeader)(unsafe.Pointer(&reasonBytes))
	context := functionContext{input: hdr}
	instance.SetContextData(unsafe.Pointer(&context))

	// Gets the `sum` exported function from the WebAssembly instance.
	contractMain, ok := instance.Exports["contract_main"]
	if !ok {
		panic("could not find function")
	}

	start := time.Now()
	for i := 0; i < 1000; i++ {
		// Calls that exported function with Go standard values. The WebAssembly
		// types are inferred and values are casted automatically.
		_, err := contractMain()
		if err != nil {
			panic(err)
		}
	}
	dur := time.Now().Sub(start)
	fmt.Printf("duration = %s\n", dur)
}

//export _reason_len
func _reason_len(context unsafe.Pointer) int32 {
	instanceContext := wasm.IntoInstanceContext(context)
	imp := (*functionContext)(instanceContext.Data())
	return imp.reasonLen(instanceContext.Memory())
}

//export _reason
func _reason(context unsafe.Pointer, ptr int32) {
	instanceContext := wasm.IntoInstanceContext(context)
	imp := (*functionContext)(instanceContext.Data())
	imp.reason(instanceContext.Memory(), ptr)
}

//export _send_transaction
func _send_transaction(context unsafe.Pointer, tagPtr, tagLen, payloadPtr, payloadLen int32) {
	instanceContext := wasm.IntoInstanceContext(context)
	imp := (*functionContext)(instanceContext.Data())
	imp.sendTransaction(instanceContext.Memory(), tagPtr, tagLen, payloadPtr, payloadLen)
}

//export _error
func _error(context unsafe.Pointer, ptr, len int32) {
	instanceContext := wasm.IntoInstanceContext(context)
	imp := (*functionContext)(instanceContext.Data())
	imp.error(instanceContext.Memory(), ptr, len)
}

type functionContext struct {
	input reflect.SliceHeader
}

func (i *functionContext) reasonLen(memory *wasm.Memory) int32 {
	log.Println("reasonLen called")
	in := *(*[]byte)(unsafe.Pointer(&i.input))
	return int32(len(in))
}

func (i *functionContext) reason(memory *wasm.Memory, ptr int32) {
	log.Println("reason called")
	data := memory.Data()
	in := *(*[]byte)(unsafe.Pointer(&i.input))
	copy(data[ptr:], in)
}

func (i *functionContext) sendTransaction(memory *wasm.Memory, tagPtr, tagLen, payloadPtr, payloadLen int32) {
	log.Println("sendTransaction called")
	data := memory.Data()
	tag := string(data[tagPtr : tagPtr+tagLen])
	payload := string(data[payloadPtr : payloadPtr+payloadLen])
	log.Printf("tag = %q; payload = %q", tag, payload)
}

func (i *functionContext) error(memory *wasm.Memory, ptr, len int32) {
	log.Println("error called")
	data := memory.Data()
	msg := string(data[ptr : ptr+len])
	log.Printf("Error: " + msg)
}
