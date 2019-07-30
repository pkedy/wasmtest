extern fn _reason_len() usize;
extern fn _reason(ptr: [*]u8) void;
extern fn _send_transaction(tag_ptr: [*]u8, tag_len: usize, payload_ptr: [*]u8, payload_len: usize) void;
extern fn _error(ptr: [*]u8, len: usize) void;

const std = @import("std");
const mem = std.mem;
const heap = std.heap;

//var arena: heap.ArenaAllocator = undefined; // = heap.ArenaAllocator.init(std.heap.wasm_allocator);
//var allocator: *mem.Allocator = undefined; // = &arena.allocator;

export fn init_mem() void {
    //arena = heap.ArenaAllocator.init(std.heap.wasm_allocator);
    //allocator = &arena.allocator;
}

const Transaction = struct {
    amount: i32,
    recipient: []i32,

    pub fn to_json(self: Transaction, a: *mem.Allocator) []u8 {
        var len: usize = 10;
        len += 14;
        for (self.recipient) |v, i| {
            len += 1;
        }
        len += 2;

        var buf = a.alloc(u8, len) catch |err| return [0]u8{};
        // TODO: Fill in buf with JSON
        mem.copy(u8, buf[0..], "TEST");
        return buf;
    }
};

export fn contract_main() void {
    //var arena = heap.ArenaAllocator.init(std.heap.wasm_allocator);
    //defer arena.deinit();
    var sbuf: [500]u8 = undefined;
    var allocator = &std.heap.FixedBufferAllocator.init(sbuf[0..]).allocator;
    //const allocator = &arena.allocator;

    //const tx = allocator.create(Transaction) catch |err| return;
    //defer heap.wasm_allocator.destroy(tx);
    //std.debug.warn("ptr={*}\n", ptr);
    //var recipient = [_]i32{1,2,3,4,5};
    //tx.amount = 12345;
    //tx.recipient = recipient[0..];

    const len = _reason_len();
    var buf = allocator.alloc(u8, len) catch |err| return;
    //defer heap.wasm_allocator.free(buf);
    _reason(buf.ptr);
    //heap.wasm_allocator.free(buf);
    const tag_value = "transfer";
    const tx = Transaction{
        .amount = 12345,
        .recipient = [_]i32{1,2,3,4,5},
    };
    const payload_value = tx.to_json(allocator);
    _send_transaction(&tag_value, tag_value.len, payload_value.ptr, payload_value.len);
    //heap.wasm_allocator.free(payload_value);
}
