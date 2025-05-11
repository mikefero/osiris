# Commands to execute which will setup the test environment

```bash
# Initial setup
koko-wd setup
deck reset -f --kong-addr http://localhost:3737/4168295f-015e-4190-837e-0fcc5d72a52f
deck sync --kong-addr http://localhost:3737/4168295f-015e-4190-837e-0fcc5d72a52f -s synthetic-no-group-association.yml

# Add consumer to consumer groups
http :3737/4168295f-015e-4190-837e-0fcc5d72a52f/consumers/ef0ce57f-00c2-446d-b5c1-fcfa1435cf91/consumer_groups \
  group=d96ddfb7-42de-4372-b9d3-d8ee854cfdd9

http :3737/4168295f-015e-4190-837e-0fcc5d72a52f/consumers/ef0ce57f-00c2-446d-b5c1-fcfa1435cf91/consumer_groups \
  group=accb1ac9-17ca-4213-be59-d1853e026404

# Update GraphQL rate limiting plugin
http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/services/d5b084da-dd01-4e61-a46b-afe265eafbd4/plugins/f958621a-e125-428b-8d03-c0c6a65cf71a \
  name=graphql-rate-limiting-advanced \
  config[window_size]:=[5] \
  config[limit]:=[30] \
  config[sync_rate]:=-1 \
  config[strategy]=redis

# Set GraphQL rate limiting costs
http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/services/d5b084da-dd01-4e61-a46b-afe265eafbd4/graphql-rate-limiting-advanced/costs/1e3ffd8b-9df4-466c-aa8a-50a008831fc9 \
  type_path=expensive \
  mul_constant=10000

# See decK issue https://github.com/Kong/deck/issues/1444 why key and key-set
# are not available in decK.
# Create key set
http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/key-sets/1dfa37e6-5e0d-4562-8578-ffe4b8787a28 \
  name=example-key-set

# Add JWK key
http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/keys/1dfa37e6-5e0d-4562-8578-ffe4b8787a28 \
  name=jwk \
  jwk='{"kty":"EC","crv":"P-256","kid":"ec-key-1","use":"sig","x":"gI0GAILBdu7T53akrFmMyGcsF3n5dO7MmwNBHKW5SV0","y":"SLW_xSffzlPWrHEVI30DHM_4egVwt3NQqeUD7nMFpps","d":"0_NxaRPUMQoAJt50Gz8YiTr8gRTwyEaCumd-MToTmIo"}' \
  kid=ec-key-1 \
  set[id]=1dfa37e6-5e0d-4562-8578-ffe4b8787a28

# Generate SSL EC key-pairs for PEM key
openssl ecparam -name prime256v1 -genkey -noout -out /tmp/ec_private_key.pem
openssl ec -in /tmp/ec_private_key.pem -pubout -out /tmp/ec_public_key.pem

# Add PEM key
http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/keys/870533a2-2616-45f0-add4-14dbf98d25b1 \
  name=pem \
  pem[private_key]=@/tmp/ec_private_key.pem \
  pem[public_key]=@/tmp/ec_public_key.pem \
  kid=ec-key-2 \
  set[id]=1dfa37e6-5e0d-4562-8578-ffe4b8787a28

# Create Redis CE partial
http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/partials/6810f599-7ae5-4178-b08a-0f7aeb3c264d \
  name=example-partial \
  type=redis-ce \
  config[host]=example.com \
  config[password]=example1234 \
  config[username]=redis

# Create Redis EE partial
http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/partials/d5daba93-1a63-4766-9f7b-522b3e8853ac \
  name=example-partial-ee \
  type=redis-ee \
  config[host]=example-ee.com \
  config[sentinel_username]=redis-ee \
  config[sentinel_password]=example-ee1234 \
  config[connection_is_proxied]:=true

# Create rate limiting plugin with partial
http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/plugins/d071c308-4bf7-4284-ae52-9aef3b6df888 \
  name=rate-limiting-advanced \
  partials:='[{"id": "d5daba93-1a63-4766-9f7b-522b3e8853ac"}]' \
  config[limit]:=[30] \
  config[window_size]:=[5]

# Create custom plugin
http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/custom-plugins/c39e3505-d2cc-4f74-a982-a0b5833ce7d8 \
  name="hello-world" \
  schema="return { name = \"hello-world\", fields = { { config = { type = \"record\", fields = { { name = { type = \"string\", required = true, }, }, }, }, }, }, }" \
  handler="local HelloWorldHandler = {} HelloWorldHandler.PRIORITY = 1000 HelloWorldHandler.VERSION = \"0.1.0\" function HelloWorldHandler:access(conf) kong.log.info(\"Hello \" .. conf.name) end return HelloWorldHandler"

# Add hello-world plugin instance
http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/plugins/fa2dba08-0e69-4cb5-a982-83d91fed1fe6 \
  name=hello-world \
  config[name]=example

# Create plugin schema
http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/v1/plugin-schemas/3462b0ed-439c-449f-af79-4fa38618825c \
  lua_schema="return { name = \"must-install\", fields = { { config = { type = \"record\", fields = { { username = { type = \"string\", required = true, default = \"chuckbilly\", }, }, }, }, }, }, }"

# Create config store and add secrets
# TODO(fero): config-stores can have the same name but different IDs; we should fix that.
CSID=$(http POST :3737/4168295f-015e-4190-837e-0fcc5d72a52f/v1/config-stores name=example-config-store | jq -r '.item.id')

http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/v1/config-stores/$CSID/secrets/example \
  value=example-value

http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/v1/config-stores/$CSID/secrets/example-2 \
  value=another-example-value

# Create vault connected to config store
http PUT :3737/4168295f-015e-4190-837e-0fcc5d72a52f/vaults/01c9e645-1fb5-4447-9da2-bfd31eae631f \
  config[config_store_id]=$CSID \
  description=konnect-config-store-vault \
  name=konnect \
  prefix=examplekonnect
```
