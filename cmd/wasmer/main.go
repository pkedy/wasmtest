package main

import (
	"encoding/json"
	"fmt"
	"log"
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

var input []byte
var instance wasm.Instance

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

	input = reasonBytes

	imp := imports{input: reasonBytes}

	// How would calling into a struct func if you want to run Wasmer in
	// a multi-threaded host?  The imports below end up invoking the exported
	// funcs instead of the ones implemented in the struct.
	imports := wasm.NewImports()
	imports.Append("_reason_len", imp.reasonLen, C._reason_len)
	imports.Append("_reason", imp.reason, C._reason)
	imports.Append("_send_transaction", imp.sendTransaction, C._send_transaction)
	imports.Append("_error", imp.error, C._error)

	// Instantiates the WebAssembly module.
	instance, err = wasm.NewInstanceWithImports(bytes, imports)
	if err != nil {
		panic(err)
	}
	defer instance.Close()

	imp.instance = &instance

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
	log.Println("_reason_len called")
	return int32(len(input))
}

//export _reason
func _reason(context unsafe.Pointer, ptr int32) {
	log.Println("_reason called")
	memory := instance.Memory.Data()
	//inst := (*wasm.Instance)(context)
	//memory := inst.Memory.Data()
	copy(memory[ptr:], input)
}

//export _send_transaction
func _send_transaction(context unsafe.Pointer, tagPtr, tagLen, payloadPtr, payloadLen int32) {
	log.Println("_send_transaction called")
	memory := instance.Memory.Data()
	tag := string(memory[tagPtr : tagPtr+tagLen])
	payload := string(memory[payloadPtr : payloadPtr+payloadLen])
	log.Printf("tag = %q; payload = %q", tag, payload)
}

//export _error
func _error(context unsafe.Pointer, ptr, len int32) {
	log.Println("_error called")
	memory := instance.Memory.Data()
	msg := string(memory[ptr : ptr+len])
	log.Printf(msg)
}

type imports struct {
	instance *wasm.Instance
	input    []byte
}

func (i *imports) reasonLen(context unsafe.Pointer) int32 {
	log.Println("reasonLen called")
	return int32(len(i.input))
}

func (i *imports) reason(context unsafe.Pointer, ptr int32) {
	log.Println("reason called")
	log.Println(ptr)
	memory := i.instance.Memory.Data()
	copy(memory[ptr:], i.input)
}

func (i *imports) sendTransaction(context unsafe.Pointer, tagPtr, tagLen, payloadPtr, payloadLen int32) {
	log.Println("sendTransaction called")
	memory := i.instance.Memory.Data()
	tag := string(memory[tagPtr : tagPtr+tagLen])
	payload := string(memory[payloadPtr : payloadPtr+payloadLen])
	log.Printf("tag = %q; payload = %q", tag, payload)
}

func (i *imports) error(context unsafe.Pointer, ptr, len int32) {
	log.Println("error called")
	memory := i.instance.Memory.Data()
	msg := string(memory[ptr : ptr+len])
	log.Printf("Error: " + msg)
}
