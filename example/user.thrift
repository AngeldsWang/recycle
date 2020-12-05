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
