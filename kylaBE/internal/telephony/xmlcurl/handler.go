// Package xmlcurl serves the FreeSWITCH mod_xml_curl protocol. FS POSTs
// requests for directory / dialplan / configuration XML and our handler
// generates the response from the Postgres-backed control plane.
//
// Auth model: the endpoint is unauthenticated at the gRPC interceptor layer
// (it lives outside the gRPC surface) but is guarded by an IP allowlist
// (the docker-compose network) and a shared secret in the X-Kyla-XML-Token
// header. FreeSWITCH config carries the token; misconfigured FS sources are
// returned a "not found" XML response.
package xmlcurl

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"kyla-be/internal/telephony"
	"kyla-be/internal/telephony/ivr"
)

// Handler bundles the stores xml-curl needs to answer FS lookups.
type Handler struct {
	store     *telephony.Store
	ivrStore  *ivr.Store
	sipRealm  string
	wssURL    string
	sharedKey string // shared secret FreeSWITCH carries in X-Kyla-XML-Token
}

func NewHandler(store *telephony.Store, ivrStore *ivr.Store, sipRealm, wssURL, sharedKey string) *Handler {
	if sipRealm == "" {
		sipRealm = "kyla"
	}
	return &Handler{
		store:     store,
		ivrStore:  ivrStore,
		sipRealm:  sipRealm,
		wssURL:    wssURL,
		sharedKey: sharedKey,
	}
}

// notFoundXML is what FreeSWITCH expects when the lookup misses. Returning
// this signals FS to fall through to its static XML / next handler.
const notFoundXML = `<?xml version="1.0" encoding="UTF-8"?>
<document type="freeswitch/xml">
  <section name="result">
    <result status="not found"/>
  </section>
</document>`

// Serve is the single Gin handler — FS posts everything to one URL and the
// `section` form field tells us what it wants.
func (h *Handler) Serve(c *gin.Context) {
	if !h.checkAuth(c) {
		c.Data(http.StatusOK, "text/xml", []byte(notFoundXML))
		return
	}
	if err := c.Request.ParseForm(); err != nil {
		log.Printf("[xmlcurl] parse form: %v", err)
		c.Data(http.StatusOK, "text/xml", []byte(notFoundXML))
		return
	}
	section := c.Request.PostFormValue("section")
	switch section {
	case "directory":
		h.handleDirectory(c, c.Request.PostForm)
	case "dialplan":
		h.handleDialplan(c, c.Request.PostForm)
	case "configuration":
		h.handleConfiguration(c, c.Request.PostForm)
	default:
		// Unknown section — let FS fall through.
		c.Data(http.StatusOK, "text/xml", []byte(notFoundXML))
	}
}

// ── auth ─────────────────────────────────────────────────────────────────────

// checkAuth gates the handler. We accept two proofs of legitimacy:
//   1. Shared-secret header (X-Kyla-XML-Token) when sharedKey is configured.
//   2. RFC1918-only source IP when no shared secret is set (dev convenience).
//
// In production both should be set.
func (h *Handler) checkAuth(c *gin.Context) bool {
	if h.sharedKey != "" {
		if c.GetHeader("X-Kyla-XML-Token") != h.sharedKey {
			log.Printf("[xmlcurl] reject: bad shared-secret header from %s", c.ClientIP())
			return false
		}
	}
	// Reject obvious public-source IPs even when the shared secret matches.
	// Defence in depth: a leaked token from a public source is still rejected.
	if !isInternalIP(c.ClientIP()) {
		log.Printf("[xmlcurl] reject: non-internal source IP %s", c.ClientIP())
		return false
	}
	return true
}

func isInternalIP(addr string) bool {
	ip := net.ParseIP(addr)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() {
		return true
	}
	return false
}

// ── directory: extension auth ──────────────────────────────────────────────

// handleDirectory answers user-lookup requests at SIP REGISTER time. FS
// supplies `key_value` (the dialled username — i.e. the extension number)
// and `domain` (the SIP realm). We return a `<user>` block with the A1 hash
// and kyla_* variables so subsequent ESL events carry org context.
func (h *Handler) handleDirectory(c *gin.Context, form url.Values) {
	extension := form.Get("user")
	if extension == "" {
		extension = form.Get("key_value")
	}
	domain := form.Get("domain")
	if extension == "" {
		c.Data(http.StatusOK, "text/xml", []byte(notFoundXML))
		return
	}
	ext, err := h.store.GetExtensionByNumber(extension)
	if err != nil {
		// Unknown extension — fall through to FS's static directory.
		c.Data(http.StatusOK, "text/xml", []byte(notFoundXML))
		return
	}
	if ext.A1Hash == "" {
		// Provisioned without an A1 hash (legacy row); deny rather than
		// allowing an unauthenticated register.
		log.Printf("[xmlcurl] directory: extension %s missing a1_hash; denying", extension)
		c.Data(http.StatusOK, "text/xml", []byte(notFoundXML))
		return
	}
	xml := renderDirectoryUser(directoryArgs{
		Domain:    domain,
		Extension: ext.Extension,
		A1Hash:    ext.A1Hash,
		OrgID:     ext.OrgID,
		UserID:    ext.UserID,
		Workspace: ext.WorkspaceID,
	})
	c.Data(http.StatusOK, "text/xml", []byte(xml))
}

// ── dialplan: inbound routing ──────────────────────────────────────────────

// handleDialplan answers per-call routing requests. FS posts the call's
// channel variables (Caller-Caller-ID-Number, Caller-Destination-Number,
// hostname, context, etc.); we look up the destination DID in our IVR
// mappings and return a `<context>` with kyla_* variables set and a `park`
// instruction so ESL takes over.
//
// When no DID mapping matches we return notFound so FS falls through to its
// static dialplan (which today is `default.xml` — also a park).
func (h *Handler) handleDialplan(c *gin.Context, form url.Values) {
	dest := form.Get("Caller-Destination-Number")
	if dest == "" {
		dest = form.Get("variable_destination_number")
	}
	context := form.Get("Hunt-Context")
	if context == "" {
		context = form.Get("context")
	}
	if context == "" {
		context = "default"
	}

	flowID, orgID, workspaceID, err := h.ivrStore.FindFlowIDForDID(dest)
	if err != nil {
		// No DID mapping — let FS fall through.
		c.Data(http.StatusOK, "text/xml", []byte(notFoundXML))
		return
	}

	xml := renderDialplanContext(dialplanArgs{
		Context:     context,
		Destination: dest,
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		IVRFlowID:   flowID,
	})
	c.Data(http.StatusOK, "text/xml", []byte(xml))
}

// ── configuration: sofia gateway provisioning ──────────────────────────────

// handleConfiguration answers module configuration lookups. We currently
// serve only "sofia.conf" — when FreeSWITCH boots (or `sofia rescan` runs)
// it asks for the full sofia configuration; we respond with a profile set
// that includes the static internal + webrtc profiles by reference (FS still
// reads those from disk first) AND an "external" profile populated with
// every active sip_trunks row as a gateway entry.
//
// Caveat: enabling dynamic sofia config means the static external profile
// XML (if any) is no longer authoritative — operators should remove
// deploy/freeswitch/conf/sip_profiles/external.xml when going live with this
// handler. See deploy/freeswitch/README.md for the migration steps.
func (h *Handler) handleConfiguration(c *gin.Context, form url.Values) {
	key := strings.TrimSpace(form.Get("key_value"))
	if key != "sofia.conf" {
		c.Data(http.StatusOK, "text/xml", []byte(notFoundXML))
		return
	}
	trunks, err := h.store.ListAllActiveTrunks()
	if err != nil {
		log.Printf("[xmlcurl] configuration: list trunks: %v", err)
		c.Data(http.StatusOK, "text/xml", []byte(notFoundXML))
		return
	}
	lite := make([]SipTrunkLite, 0, len(trunks))
	for _, t := range trunks {
		lite = append(lite, SipTrunkLite{
			GatewayName: t.GatewayName,
			Username:    t.Username,
			Password:    t.Password,
			SipServer:   t.SipServer,
			FromURI:     t.FromURI,
		})
	}
	xml := renderSofiaConfiguration(lite)
	c.Data(http.StatusOK, "text/xml", []byte(xml))
}
