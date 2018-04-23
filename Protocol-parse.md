List of op-codes and req-res to device with my comments.
Useful links:
https://medium.freecodecamp.org/how-i-hacked-xiaomi-miband-2-to-control-it-from-linux-a5bd2f36d3ad

# OP-codes

Opcode: Read By Type Request (`0x08`)

    0... .... = Authentication Signature: False
    .0.. .... = Command: False
    ..00 1000 = Method: Read By Type Request (0x08)

Opcode: Read By Type Response (`0x09`)

    0... .... = Authentication Signature: False
    .0.. .... = Command: False
    ..00 1001 = Method: Read By Type Response (0x09)


Opcode: Write Command (`0x52`)

    0... .... = Authentication Signature: False
    .1.. .... = Command: True
    ..01 0010 = Method: Write Request (0x12)


Opcode: Read Request (`0x0a`)

    0... .... = Authentication Signature: False
    .0.. .... = Command: False
    ..00 1010 = Method: Read Request (0x0a)


Opcode: Read Response (`0x0b`)

    0... .... = Authentication Signature: False
    .0.. .... = Command: False
    ..00 1011 = Method: Read Response (0x0b)


Opcode: Handle Value Notification (`0x1b`)

    0... .... = Authentication Signature: False
    .0.. .... = Command: False
    ..01 1011 = Method: Handle Value Notification (0x1b)


# Request-response
Sent: Get the device name

    Opcode: Read By Type Request (0x08)
    Starting Handle: 0x0001
    Ending Handle: 0xffff
    UUID: Device Name (0x2a00)

Rcvd: Return the device name: `4d492042616e642032` > `MI Band 2`

    Opcode: Read By Type Response (0x09)
    Length: 11 (0x0b)
    Attribute Data, Handle: 0x0003
        Handle: 0x0003: Device Name)
        Device Name: MI Band 2 (4d 49 20 42 61 6e 64 20 32)
    [UUID: Device Name (0x2a00)]

Sent: Auth connection request

    Opcode: Write Command (0x52)
    Handle: 0x0055 (Unknown)
    Value: 0100

Sent: Pass 16 bytes aes key

    Opcode: Write Command (0x52)
    Handle: 0x0054 (Unknown)
    Value: 0100 d88ca0fef504cc82735f15feb6bf9b1f
    Value: 0100 3fa88ed11d6f574047924eaa3a060fb2

Rcvd: Paired successfully (manual pairing by typing on band)

    Opcode: Handle Value Notification (0x1b)
    Handle: 0x0054 (Unknown)
    Value: 100101

Sent: Requesting random key from the device

    Opcode: Write Command (0x52)
    Handle: 0x0054 (Unknown)
    Value: 0200

Rcvd: Getting random number from the device (last 16 bytes)

    Opcode: Handle Value Notification (0x1b)
    Handle: 0x0054 (Unknown)
    Value: 100201 689c1dd58a666f0f8c90a1a96e1ab3a3

Sent: Send back random number aes-encrypted with own key

    Opcode: Write Command (0x52)
    Handle: 0x0054 (Unknown)
    Value: 0300 ce1c0135db39e61f1a27012fb907df3b

Rcvd: Assume receive confirmation of success pairing

    Opcode: Handle Value Notification (0x1b)
    Handle: 0x0054 (Unknown)
    Value: 100301

<======= I'm here right now in my code.

Sent: Another command request

    Opcode: Write Command (0x52)
    Handle: 0x0055 (Unknown)
    Value: 0000

Sent: Read value request

    Opcode: Read Request (0x0a)
    Handle: 0x002f (Unknown)

Rcvd: Read value response

    Opcode: Read Response (0x0b)
    [Handle: 0x002f (Unknown)]
    Value: e207040f10142b0700000c

Sent: The same, but why? Probably to confirm

    Opcode: Write Request (0x12)
    Handle: 0x002f (Unknown)
    Value: e207040f10142f0000000c

Rcvd: And again, only Opcode is different

    Opcode: Write Response (0x13)
    [Handle: 0x002f (Unknown)]

Sent: Have no idea

    Opcode: Read Request (0x0a)
    Handle: 0x0012 (Unknown)

Rcvd: Stuck here

    Opcode: Read Response (0x0b)
    [Handle: 0x0012 (Unknown)]
    Value: 56312e302e312e3831

I think that's enough.