local key = KEYS[1]
--- 用户输入的 code
local expectedCode = ARGV[1]
local cnt = tonumber(redis.call("get", key .. ":cnt"))
local code = redis.call("get", key)

if cnt <= 0 then
    -- 说明 用户一直输入错误
    return -1
elseif expectedCode == code then
    redis.call("del", key)
    return 0
else
    redis.call("decr", key .. ":cnt" )
    return -2
end
