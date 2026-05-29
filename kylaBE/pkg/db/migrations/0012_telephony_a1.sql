-- Phase 5 — SIP digest auth (mod_xml_curl directory lookup support)
-- Adds the HA1 hash column to sip_extensions. mod_xml_curl serves SIP
-- registrations: FreeSWITCH expects an HA1 digest (md5(user:realm:password))
-- rather than a bcrypt hash, which can't be inverted to compute the digest.
--
-- bcrypt's password_hash column is kept for now to track historical state but
-- the SIP auth path uses a1_hash exclusively.

ALTER TABLE sip_extensions
  ADD COLUMN IF NOT EXISTS a1_hash TEXT NOT NULL DEFAULT '';

-- We don't backfill: existing extensions will need to be re-provisioned via
-- CreateSipExtension to populate a1_hash. That re-provisioning issues a new
-- SIP password and pushes it to the PBX via ProvisionExtension.
