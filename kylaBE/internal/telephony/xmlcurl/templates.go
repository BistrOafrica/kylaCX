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

// esc XML-escapes a string. html.EscapeString covers the same set of chars
// XML requires (&, <, >, ", ').
func esc(s string) string {
	return html.EscapeString(s)
}
