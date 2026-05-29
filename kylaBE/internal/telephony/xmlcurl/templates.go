package xmlcurl

import (
	"fmt"
	"html"
	"strings"
)

// directoryArgs is the input to renderDirectoryUser.
type directoryArgs struct {
	Domain    string
	Extension string
	A1Hash    string
	OrgID     string
	UserID    string
	Workspace string
}

// renderDirectoryUser returns the FreeSWITCH directory XML for one user. The
// `params.a1-hash` is the SIP digest secret; FS uses it to challenge the
// REGISTER request without needing the plaintext password.
//
// The `<variables>` block stamps kyla_* channel variables on every leg this
// user originates so ESL events downstream carry org/workspace context
// without an additional DB lookup.
func renderDirectoryUser(a directoryArgs) string {
	domain := a.Domain
	if domain == "" {
		domain = "kyla"
	}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<document type="freeswitch/xml">` + "\n")
	b.WriteString(`  <section name="directory">` + "\n")
	fmt.Fprintf(&b, `    <domain name="%s">`+"\n", esc(domain))
	b.WriteString(`      <users>` + "\n")
	fmt.Fprintf(&b, `        <user id="%s">`+"\n", esc(a.Extension))
	b.WriteString(`          <params>` + "\n")
	fmt.Fprintf(&b, `            <param name="a1-hash" value="%s"/>`+"\n", esc(a.A1Hash))
	b.WriteString(`          </params>` + "\n")
	b.WriteString(`          <variables>` + "\n")
	fmt.Fprintf(&b, `            <variable name="user_context" value="default"/>`+"\n")
	fmt.Fprintf(&b, `            <variable name="kyla_org_id" value="%s"/>`+"\n", esc(a.OrgID))
	fmt.Fprintf(&b, `            <variable name="kyla_user_id" value="%s"/>`+"\n", esc(a.UserID))
	fmt.Fprintf(&b, `            <variable name="kyla_workspace_id" value="%s"/>`+"\n", esc(a.Workspace))
	b.WriteString(`          </variables>` + "\n")
	b.WriteString(`        </user>` + "\n")
	b.WriteString(`      </users>` + "\n")
	b.WriteString(`    </domain>` + "\n")
	b.WriteString(`  </section>` + "\n")
	b.WriteString(`</document>` + "\n")
	return b.String()
}

// dialplanArgs is the input to renderDialplanContext.
type dialplanArgs struct {
	Context     string
	Destination string
	OrgID       string
	WorkspaceID string
	IVRFlowID   string
}

// renderDialplanContext returns a dialplan extension that sets kyla_*
// channel variables on the inbound leg and parks the call so ESL can take
// over routing. The EventBridge's onCreate will see the kyla_org_id stamped
// here and skip its own DID lookup.
func renderDialplanContext(a dialplanArgs) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<document type="freeswitch/xml">` + "\n")
	b.WriteString(`  <section name="dialplan">` + "\n")
	fmt.Fprintf(&b, `    <context name="%s">`+"\n", esc(a.Context))
	fmt.Fprintf(&b, `      <extension name="kyla_inbound_%s">`+"\n", esc(a.Destination))
	fmt.Fprintf(&b, `        <condition field="destination_number" expression="^%s$">`+"\n", esc(a.Destination))
	fmt.Fprintf(&b, `          <action application="set" data="kyla_org_id=%s"/>`+"\n", esc(a.OrgID))
	fmt.Fprintf(&b, `          <action application="set" data="kyla_workspace_id=%s"/>`+"\n", esc(a.WorkspaceID))
	fmt.Fprintf(&b, `          <action application="set" data="kyla_ivr_flow_id=%s"/>`+"\n", esc(a.IVRFlowID))
	b.WriteString(`          <action application="set" data="kyla_did=${destination_number}"/>` + "\n")
	b.WriteString(`          <action application="park"/>` + "\n")
	b.WriteString(`        </condition>` + "\n")
	b.WriteString(`      </extension>` + "\n")
	b.WriteString(`    </context>` + "\n")
	b.WriteString(`  </section>` + "\n")
	b.WriteString(`</document>` + "\n")
	return b.String()
}

// renderSofiaConfiguration returns a full sofia.conf XML with the supplied
// trunks installed as gateways on the "external" profile. The internal and
// webrtc profiles are reserved for static config (operators keep them under
// sip_profiles/) so this fragment only re-creates external.
//
// FreeSWITCH expects the configuration response to be a complete document
// rooted at <document type="freeswitch/xml">; we wrap accordingly.
func renderSofiaConfiguration(trunks []SipTrunkLite) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<document type="freeswitch/xml">` + "\n")
	b.WriteString(`  <section name="configuration">` + "\n")
	b.WriteString(`    <configuration name="sofia.conf" description="sofia Endpoint">` + "\n")
	b.WriteString(`      <global_settings>` + "\n")
	b.WriteString(`        <param name="log-level" value="0"/>` + "\n")
	b.WriteString(`        <param name="auto-restart" value="false"/>` + "\n")
	b.WriteString(`        <param name="debug-presence" value="0"/>` + "\n")
	b.WriteString(`      </global_settings>` + "\n")
	b.WriteString(`      <profiles>` + "\n")
	b.WriteString(`        <profile name="external">` + "\n")
	b.WriteString(`          <gateways>` + "\n")
	for _, t := range trunks {
		fmt.Fprintf(&b, `            <gateway name="%s">`+"\n", esc(t.GatewayName))
		fmt.Fprintf(&b, `              <param name="username" value="%s"/>`+"\n", esc(t.Username))
		fmt.Fprintf(&b, `              <param name="password" value="%s"/>`+"\n", esc(t.Password))
		fmt.Fprintf(&b, `              <param name="realm" value="%s"/>`+"\n", esc(t.SipServer))
		fmt.Fprintf(&b, `              <param name="proxy" value="%s"/>`+"\n", esc(t.SipServer))
		if t.FromURI != "" {
			fmt.Fprintf(&b, `              <param name="from-domain" value="%s"/>`+"\n", esc(t.FromURI))
		}
		// register=true means we maintain SIP registration with the provider —
		// most trunks expect this; register-by-IP providers can override the
		// trunk row to set username/password empty.
		fmt.Fprintf(&b, `              <param name="register" value="%s"/>`+"\n", boolXMLAttr(t.Username != ""))
		b.WriteString(`              <param name="caller-id-in-from" value="true"/>` + "\n")
		b.WriteString(`            </gateway>` + "\n")
	}
	b.WriteString(`          </gateways>` + "\n")
	b.WriteString(`          <settings>` + "\n")
	b.WriteString(`            <param name="sip-port" value="5080"/>` + "\n")
	b.WriteString(`            <param name="rtp-ip" value="$${local_ip_v4}"/>` + "\n")
	b.WriteString(`            <param name="sip-ip" value="$${local_ip_v4}"/>` + "\n")
	b.WriteString(`            <param name="ext-rtp-ip" value="auto-nat"/>` + "\n")
	b.WriteString(`            <param name="ext-sip-ip" value="auto-nat"/>` + "\n")
	b.WriteString(`            <param name="auth-calls" value="false"/>` + "\n")
	b.WriteString(`            <param name="context" value="public"/>` + "\n")
	b.WriteString(`            <param name="dialplan" value="XML"/>` + "\n")
	b.WriteString(`            <param name="dtmf-duration" value="2000"/>` + "\n")
	b.WriteString(`            <param name="dtmf-type" value="rfc2833"/>` + "\n")
	b.WriteString(`            <param name="inbound-codec-prefs" value="$${global_codec_prefs}"/>` + "\n")
	b.WriteString(`            <param name="outbound-codec-prefs" value="$${outbound_codec_prefs}"/>` + "\n")
	b.WriteString(`            <param name="user-agent-string" value="kyla-pbx-external"/>` + "\n")
	b.WriteString(`          </settings>` + "\n")
	b.WriteString(`        </profile>` + "\n")
	b.WriteString(`      </profiles>` + "\n")
	b.WriteString(`    </configuration>` + "\n")
	b.WriteString(`  </section>` + "\n")
	b.WriteString(`</document>` + "\n")
	return b.String()
}

// SipTrunkLite is the subset of telephony.SipTrunk the templates need.
// Declared locally so this package doesn't have to import internal/telephony
// for type-only purposes (the handler already does — it constructs the slice
// from the full type via TrunksToLite below).
type SipTrunkLite struct {
	GatewayName string
	Username    string
	Password    string
	SipServer   string
	FromURI     string
}

func boolXMLAttr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// esc XML-escapes a string. html.EscapeString covers the same set of chars
// XML requires (&, <, >, ", ').
func esc(s string) string {
	return html.EscapeString(s)
}
