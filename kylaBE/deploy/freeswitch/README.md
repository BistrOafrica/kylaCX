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

## mod_xml_curl integration (live as of 2026-05-29)

FreeSWITCH consults the Go backend over HTTP for **directory** and **dialplan** XML:

- `directory`: at SIP REGISTER, FS posts the extension number → backend returns the user XML with the A1 digest hash + kyla_* channel variables.
- `dialplan`: at CHANNEL_CREATE, FS posts the destination → backend looks up the DID in `ivr_did_mappings` and returns a context that stamps `kyla_org_id` / `kyla_workspace_id` / `kyla_ivr_flow_id` and parks for ESL.
- `configuration`: currently always returns `not found`; reserved for future trunk provisioning over XML.

Wiring lives in `autoload_configs/xml_curl.conf.xml`. **Replace the placeholder `CHANGE_ME_TOKEN`** with the value of the backend's `FS_XML_CURL_TOKEN` env var before exposing the stack beyond local dev. Without a token match the handler returns `not found`, forcing FS to fall through to its static XML.

The handler enforces an RFC1918 source-IP allowlist as defence in depth, so a leaked token from a public IP is still rejected.

## Dynamic sofia gateways (live as of slice 5 polish)

The `configuration` section of mod_xml_curl now serves a complete `sofia.conf` populated with every `sip_trunks` row (where `is_active = true`) as a gateway on the **external** profile.

To enable:

1. Remove (or rename) `sip_profiles/external.xml` so FreeSWITCH doesn't load a static external profile that conflicts with the dynamic one.
2. Restart FreeSWITCH (or run `reload mod_sofia` from `fs_cli`).
3. CRUD via the SIP admin UI now propagates to the PBX on the next sofia reload — after editing a trunk, run `sofia profile external rescan` from `fs_cli` to pick up the new gateway.

The internal + webrtc profiles are deliberately **not** served by xml-curl. They stay statically configured because their behaviour rarely changes per-tenant. The external profile is the one most likely to need tenant-specific gateways.

## What's deliberately deferred

1. **TLS certs for WSS** — the dev image uses a self-signed certificate. For production replace `/etc/freeswitch/tls/wss.pem` with a real cert (Let's Encrypt fullchain + privkey concatenated).

2. **Per-tenant PBX instances** — the dynamic sofia config serves trunks from every org under a single external profile, scoped by gateway name. Genuine tenant isolation requires one FreeSWITCH per org or a per-profile xml-curl binding.

3. **Legacy static directory** — the `directory/default/*.xml` path is still searched by FS before consulting mod_xml_curl. Empty by default; remove unused entries to avoid surprise authorisations.

4. **Live trunk reload via gRPC** — `UpdateSipTrunk` writes to Postgres but doesn't trigger `sofia profile external rescan` automatically. Operators must reload manually or wait for the next FS restart.

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
