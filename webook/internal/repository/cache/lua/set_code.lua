-- 验证码在 redis 上的 key
local key = keys[1]
-- 一个验证码验证的次数
local cntKey = key .. ":cnt"
-- 你的验证码
local val = ARGV[1]

local ttl = tonumber(redis.call("ttl", key))

if ttl == -1 then
    -- key 存在但是没有过期时间
    return -2 -- 系统错误
    -- -2 是 key 不存在, ttl < 540 是已经发送了一个验证码超过了一分钟
elseif ttl == -2 or ttl < 540 then
    redis.call("set", key, val)
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)
    redis.call("expire",cntkey, 600)
    return 0
else
    return -1 -- 发送太频繁
end

