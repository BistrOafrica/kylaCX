# FreeSWITCH config skeleton

Minimal config that boots FreeSWITCH with everything kyla's Go telephony service expects:

- ESL bind on `0.0.0.0:8021` (password from `default_password` in `vars.xml`)
- SIP/UDP profile on 5060 for SIP desk phones and on-net extensions
- SIP-over-WSS profile on 7443 for browser softphones (WebRTC + OPUS + DTLS-SRTP)
- A default dialplan that parks calls so ESL can take over routing
- mod_say_en for IVR `say` nodes; mod_dptools for `play_and_get_digits`

## Mount layout

`docker-compose.yaml` mounts this directory at `/etc/freeswitch` inside the container:

```
deploy/freeswitch/conf/
  vars.xml
  autoload_configs/
    event_socket.conf.xml
    modules.conf.xml
  sip_profiles/
    internal.xml
    webrtc.xml
  dialplan/
    default.xml
```

## What's deliberately deferred

1. **mod_xml_curl integration** — the Go backend doesn't yet serve a directory or dialplan over HTTP. Today, SIP extensions must be provisioned statically in `directory/default/*.xml`. The roadmap follow-up: implement a Gin handler at `/freeswitch/xml-curl` that returns directory / dialplan / config XML on demand from the Postgres `sip_extensions`, `sip_trunks`, and `ivr_did_mappings` tables.

2. **TLS certs for WSS** — the dev image uses a self-signed certificate. For production replace `/etc/freeswitch/tls/wss.pem` with a real cert (Let's Encrypt fullchain + privkey concatenated).

3. **Sofia gateway profiles (PSTN trunks)** — `sip_trunks` rows hold credentials but the Go service's `ProvisionTrunk` is a no-op today. Until mod_xml_curl is wired, gateways must be added under `sip_profiles/external/<trunk_name>.xml` manually.

4. **Inbound DID-to-org mapping at the dialplan layer** — for now the EventBridge looks up the DID in the `ivr_did_mappings` table at CHANNEL_CREATE time. mod_xml_curl will eventually let the dialplan stamp `kyla_org_id` directly so inbound non-IVR calls also carry org context.

## Verifying the stack

After `docker compose -f deploy/docker-compose.yaml up`:

```bash
# ESL connectivity from the kyla backend container:
docker compose exec grpc-server nc -v freeswitch 8021

# Echo test from a softphone — dial 9999.
# Expect bidirectional audio.

# Tail FS logs:
docker compose logs -f freeswitch
```

The Go telephony service's startup log will print `freeswitch: connected to ESL freeswitch:8021 at ...` once the ESL handshake succeeds.
