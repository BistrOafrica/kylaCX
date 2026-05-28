import { describe, it, expect } from "vitest"
import { toRecord, TYPE_SLUG } from "@/features/crm/utils/types"
import { Object as CoreObject } from "@/pb/object_core"

describe("toRecord", () => {
  it("parses the JSON data blob into an object", () => {
    const obj = CoreObject.create({
      id: "obj-1",
      orgId: "org-1",
      workspaceId: "ws-1",
      typeSlug: "contact",
      data: '{"name":"Maria","email":"m@k.io"}',
      createdBy: "u-1",
      createdAt: "2026-05-28T12:00:00Z",
      updatedAt: "2026-05-28T12:00:00Z",
    }) as CoreObject

    const r = toRecord(obj)
    expect(r.id).toBe("obj-1")
    expect(r.data.name).toBe("Maria")
    expect(r.data.email).toBe("m@k.io")
  })

  it("returns empty data when JSON is malformed", () => {
    const obj = CoreObject.create({
      id: "obj-2",
      orgId: "org-1",
      workspaceId: "ws-1",
      typeSlug: "contact",
      data: "not-json",
      createdBy: "u-1",
      createdAt: "",
      updatedAt: "",
    }) as CoreObject
    expect(toRecord(obj).data).toEqual({})
  })

  it("returns empty data when blob is empty", () => {
    const obj = CoreObject.create({
      id: "obj-3",
      orgId: "org-1",
      workspaceId: "ws-1",
      typeSlug: "contact",
      data: "",
      createdBy: "",
      createdAt: "",
      updatedAt: "",
    }) as CoreObject
    expect(toRecord(obj).data).toEqual({})
  })
})

describe("TYPE_SLUG", () => {
  it("contains the canonical CRM type slugs", () => {
    expect(TYPE_SLUG.contact).toBe("contact")
    expect(TYPE_SLUG.deal).toBe("deal")
    expect(TYPE_SLUG.company).toBe("company")
    expect(TYPE_SLUG.lead).toBe("lead")
  })
})
