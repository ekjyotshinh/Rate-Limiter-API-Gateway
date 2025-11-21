-- KEYS[1] = key (string)
-- ARGV[1] = now_ms (number)
-- ARGV[2] = capacity (number)
-- ARGV[3] = refill_rate (tokens per second, number)
-- ARGV[4] = cost (number)
-- ARGV[5] = ttl_seconds (number)

local key = KEYS[1]
local now = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local refill_rate = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])
local ttl = tonumber(ARGV[5])

local data = redis.call("HMGET", key, "tokens", "last")
local tokens = tonumber(data[1])
local last = tonumber(data[2])

if tokens == nil or last == nil then
  tokens = capacity
  last = now
end

local elapsed = (now - last) / 1000.0
local refill = elapsed * refill_rate
tokens = math.min(capacity, tokens + refill)
last = now

local allowed = 0
if tokens >= cost then
  tokens = tokens - cost
  allowed = 1
end

redis.call("HMSET", key, "tokens", tostring(tokens), "last", tostring(last))
redis.call("EXPIRE", key, ttl)

-- return allowed (0/1), tokens_left (string), capacity (string)
return {allowed, tostring(tokens), tostring(capacity)}
