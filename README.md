# Recycle
Recycle works like a `thrift store` polishing bytes to original shape.

## Usage

``` shell
→ ./recycle -h
Given the bytes encoded by thrift protocols and the target definition in thrift IDL,
recycle can restore the data with specific type names instead of field numbers.

Usage:
  recycle [flags]

Flags:
  -h, --help            help for recycle
  -f, --thrift string   thrift idl file path
  -t, --type string     target type name in thrift idl
```

For example, you have a rpc service for creating user.
``` shell
$ cat ./example/user.thrift
namespace go example.recycle

struct Profile {
    1: list<string> Interests
    2: i32          Age
}

struct CreateUserRequest {
    1:  required string Name
    10: optional Profile Profile
}

struct CreateUserResponse {
    1: i64 ID
}

service UserService {
    CreateUserResponse CreateUser(1: CreateUserRequest req)
}

```

Assume you captured the byte stream for a specific rpc request and dumped them with base64 encoded string, you can restore the `CreateUserRequest` with a readable json format.
``` shell
$ echo 'gAEAAQAAAApDcmVhdGVVc2VyAAAAewwAAQwACg8AAQoAAAAACAACAAAAFwALAAEAAAAGdGhyaWZ0AAA=' \
| ./recycle  -f ./example/user.thrift -t CreateUserRequest | jq
{
  "Name": "thrift",
  "Profile": {
    "Age": 23,
    "Interests": []
  }
}
```
