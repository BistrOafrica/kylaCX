import { useCallback, useEffect, useMemo, useState } from "react"
import {
  ReactFlow,
  Background,
  Controls,
  Handle,
  Position,
  type Node,
  type Edge,
  type NodeChange,
  type EdgeChange,
  type Connection,
  applyNodeChanges,
  applyEdgeChanges,
  addEdge,
} from "@xyflow/react"
import "@xyflow/react/dist/style.css"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import {
  IconArrowLeft,
  IconDeviceFloppy,
  IconLoader2,
  IconPlus,
  IconPlayerPlay,
  IconMessage2,
  IconArrowsRightLeft,
  IconHandStop,
  IconRecordMail,
  IconMusic,
  IconArrowRightCircle,
} from "@tabler/icons-react"
import { Link } from "react-router-dom"
import { toast } from "sonner"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import type { IVRFlow, IVRNode } from "@/pb/ivr"
import { Struct } from "@/pb/google/protobuf/struct"
import {
  getIvrV2Flow,
  updateIvrV2Flow,
  nodeConfig,
  setNodeConfig,
} from "../api/ivrV2"

/**
 * IvrV2Builder — the visual flow editor targeting the Phase 5c IVRService.
 *
 * Layout: left palette (drag/click to add nodes), centre React Flow canvas,
 * right-side inspector for the selected node's config. Saving serialises the
 * canvas state back into IVRFlow.definition.{startNodeId,nodes} and POSTs via
 * the new gRPC service.
 *
 * Node model: each canvas node carries one IVRNode in its `data` payload.
 * Menu nodes render an extra outgoing handle per `branches` key (digit) so
 * the user can wire branch-specific transitions visually.
 *
 * Edge model: edges represent next-node transitions. Edges without a
 * sourceHandle map to node.next_node_id; edges with a sourceHandle named
 * after a digit map to node.branches[digit].
 */

interface BuilderProps {
  id: string
}

// Node type metadata used by the palette + custom rendering.
const NODE_TYPES = [
  { key: "play_audio", label: "Play audio", icon: IconMusic, palette: "Plays a static audio file" },
  { key: "say", label: "Say (TTS)", icon: IconMessage2, palette: "Synthesise text via TTS" },
  { key: "menu", label: "Menu (DTMF)", icon: IconArrowsRightLeft, palette: "Collect a digit, branch on it" },
  { key: "transfer", label: "Transfer", icon: IconArrowRightCircle, palette: "Hand the call to an extension/number" },
  { key: "record", label: "Record", icon: IconRecordMail, palette: "Record audio to disk" },
  { key: "hangup", label: "Hang up", icon: IconHandStop, palette: "End the call" },
  { key: "goto", label: "Go to", icon: IconPlayerPlay, palette: "Jump to another node" },
] as const

type NodeTypeKey = (typeof NODE_TYPES)[number]["key"]

export function IvrV2Builder({ id }: BuilderProps) {
  const qc = useQueryClient()

  const flowQuery = useQuery({
    queryKey: ["ivr-v2-flow", id],
    queryFn: () => getIvrV2Flow(id),
  })

  // React Flow's controlled state. We hydrate from the loaded flow on first
  // render, then own the canvas state until save.
  const [nodes, setNodes] = useState<Node[]>([])
  const [edges, setEdges] = useState<Edge[]>([])
  const [name, setName] = useState("")
  const [startNodeId, setStartNodeId] = useState<string>("")
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)

  // Hydrate canvas when the flow loads.
  useEffect(() => {
    const data = flowQuery.data
    if (!data) return
    setName(data.name ?? "")
    const def = data.definition
    if (!def) return
    setStartNodeId(def.startNodeId ?? "")
    const initialNodes: Node[] = (def.nodes ?? []).map((n, i) => ({
      id: n.id,
      type: "ivr",
      position: positionForIndex(i),
      data: { node: n, isStart: n.id === def.startNodeId },
    }))
    const initialEdges: Edge[] = []
    for (const n of def.nodes ?? []) {
      if (n.nextNodeId) {
        initialEdges.push({ id: `${n.id}->${n.nextNodeId}`, source: n.id, target: n.nextNodeId })
      }
      for (const [digit, target] of Object.entries(n.branches ?? {})) {
        initialEdges.push({
          id: `${n.id}-${digit}->${target}`,
          source: n.id,
          sourceHandle: digit,
          target,
          label: digit,
        })
      }
    }
    setNodes(initialNodes)
    setEdges(initialEdges)
  }, [flowQuery.data])

  // React Flow change handlers.
  const onNodesChange = useCallback(
    (changes: NodeChange[]) => setNodes((ns) => applyNodeChanges(changes, ns)),
    [],
  )
  const onEdgesChange = useCallback(
    (changes: EdgeChange[]) => setEdges((es) => applyEdgeChanges(changes, es)),
    [],
  )
  const onConnect = useCallback(
    (conn: Connection) => setEdges((es) => addEdge({ ...conn, label: conn.sourceHandle ?? undefined }, es)),
    [],
  )

  // Add a new node from the palette. Drops it in the centre of the viewport.
  const addNode = useCallback(
    (type: NodeTypeKey) => {
      const newId = `n_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 6)}`
      const node: IVRNode = {
        id: newId,
        type,
        config: undefined,
        nextNodeId: "",
        branches: {},
      } as IVRNode
      setNodes((ns) => [
        ...ns,
        {
          id: newId,
          type: "ivr",
          position: { x: 200 + ns.length * 40, y: 100 + ns.length * 40 },
          data: { node, isStart: ns.length === 0 && startNodeId === "" },
        },
      ])
      if (!startNodeId && nodes.length === 0) setStartNodeId(newId)
      setSelectedNodeId(newId)
    },
    [nodes.length, startNodeId],
  )

  // Save: serialise canvas state back into IVRFlow.definition and PUT.
  const save = useMutation({
    mutationFn: async () => {
      if (!flowQuery.data) throw new Error("flow not loaded")
      const serialisedNodes = nodes.map((rn) => {
        const node = rn.data.node as IVRNode
        // Reset routing fields; we recompute them from edges.
        node.nextNodeId = ""
        node.branches = {}
        return node
      })
      // Apply edges back onto nodes.
      for (const e of edges) {
        const src = serialisedNodes.find((n) => n.id === e.source)
        if (!src) continue
        if (e.sourceHandle) {
          src.branches = { ...(src.branches ?? {}), [e.sourceHandle]: e.target }
        } else {
          src.nextNodeId = e.target
        }
      }
      const updated: IVRFlow = {
        ...flowQuery.data,
        name,
        definition: {
          startNodeId,
          nodes: serialisedNodes,
        },
      }
      return updateIvrV2Flow(updated)
    },
    onSuccess: () => {
      toast.success("Flow saved")
      qc.invalidateQueries({ queryKey: ["ivr-v2-flow", id] })
    },
    onError: (e: Error) => toast.error(e.message),
  })

  const customNodeTypes = useMemo(() => ({ ivr: IvrCanvasNode }), [])

  if (flowQuery.isPending) {
    return (
      <div className="p-6 text-fg-muted">
        <IconLoader2 className="size-5 animate-spin" /> Loading flow…
      </div>
    )
  }
  if (flowQuery.isError) {
    return <div className="p-6 text-danger">Couldn't load flow: {(flowQuery.error as Error).message}</div>
  }

  const selectedNode = nodes.find((n) => n.id === selectedNodeId)
  return (
    <div className="flex flex-col h-full">
      <header className="flex items-center gap-3 p-3 border-b">
        <Link to="/calls/ivr" className="text-fg-muted hover:text-fg">
          <IconArrowLeft className="size-5" />
        </Link>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Flow name"
          className="max-w-sm"
        />
        <button
          onClick={() => save.mutate()}
          disabled={save.isPending}
          className={cn(
            "ms-auto inline-flex items-center gap-1.5 h-8 px-3 rounded-sm",
            "bg-accent text-accent-fg hover:opacity-90",
            "disabled:opacity-50 disabled:pointer-events-none",
          )}
        >
          {save.isPending ? (
            <IconLoader2 className="size-4 animate-spin" />
          ) : (
            <IconDeviceFloppy className="size-4" />
          )}
          Save
        </button>
      </header>

      <div className="flex flex-1 min-h-0">
        {/* Left palette */}
        <aside className="w-56 border-e p-3 overflow-y-auto space-y-1">
          <div className="text-xs uppercase font-medium text-fg-muted px-1 pb-1">Nodes</div>
          {NODE_TYPES.map((t) => (
            <button
              key={t.key}
              onClick={() => addNode(t.key)}
              className="w-full flex items-start gap-2 p-2 rounded-sm hover:bg-subtle text-left"
            >
              <t.icon className="size-4 mt-0.5 text-fg-muted" />
              <div className="flex-1">
                <div className="text-sm font-medium">{t.label}</div>
                <div className="text-xs text-fg-muted">{t.palette}</div>
              </div>
              <IconPlus className="size-3.5 text-fg-muted" />
            </button>
          ))}
        </aside>

        {/* Canvas */}
        <div className="flex-1 min-w-0 bg-canvas">
          <ReactFlow
            nodes={nodes.map((n) => ({
              ...n,
              data: { ...n.data, isStart: n.id === startNodeId },
            }))}
            edges={edges}
            nodeTypes={customNodeTypes}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={(_, n) => setSelectedNodeId(n.id)}
            fitView
          >
            <Background />
            <Controls />
          </ReactFlow>
        </div>

        {/* Inspector */}
        <aside className="w-72 border-s p-3 overflow-y-auto">
          {selectedNode ? (
            <NodeInspector
              node={selectedNode.data.node as IVRNode}
              isStart={selectedNode.id === startNodeId}
              onMarkStart={() => setStartNodeId(selectedNode.id)}
              onChange={(updated) => {
                setNodes((ns) =>
                  ns.map((n) =>
                    n.id === selectedNode.id ? { ...n, data: { ...n.data, node: updated } } : n,
                  ),
                )
              }}
            />
          ) : (
            <div className="text-sm text-fg-muted">Select a node to edit.</div>
          )}
        </aside>
      </div>
    </div>
  )
}

// ── Custom canvas node ─────────────────────────────────────────────────────

interface CanvasNodeData {
  node: IVRNode
  isStart: boolean
}

// IvrCanvasNode renders a node card with type-aware handles. Menu nodes get
// one output handle per branch digit so connections route by DTMF.
function IvrCanvasNode({ data }: { data: CanvasNodeData }) {
  const { node, isStart } = data
  const meta = NODE_TYPES.find((t) => t.key === node.type)
  const Icon = meta?.icon ?? IconMessage2
  const branchKeys = Object.keys(node.branches ?? {}).sort()
  return (
    <div
      className={cn(
        "min-w-[160px] rounded-md border bg-bg shadow-sm",
        isStart ? "border-success" : "border-default",
      )}
    >
      <Handle type="target" position={Position.Top} />
      <div className="flex items-center gap-2 p-2 border-b">
        <Icon className="size-4 text-fg-muted" />
        <div className="text-xs font-medium uppercase">{meta?.label ?? node.type}</div>
        {isStart && <span className="ms-auto text-[10px] font-bold text-success">START</span>}
      </div>
      <div className="p-2 text-xs text-fg-muted">
        {summariseNode(node)}
      </div>
      {/* Default outgoing handle (next_node_id). */}
      {node.type !== "menu" && node.type !== "hangup" && (
        <Handle type="source" position={Position.Bottom} />
      )}
      {/* Menu nodes: one handle per branch digit, plus a default. */}
      {node.type === "menu" && (
        <div className="flex flex-wrap items-center gap-2 p-2 border-t">
          {branchKeys.length === 0 ? (
            <div className="text-[10px] text-fg-muted">Add branch keys in inspector →</div>
          ) : (
            branchKeys.map((k) => (
              <div key={k} className="relative px-1.5 py-0.5 text-[10px] rounded-xs bg-subtle">
                {k}
                <Handle
                  id={k}
                  type="source"
                  position={Position.Bottom}
                  style={{ left: "50%" }}
                />
              </div>
            ))
          )}
        </div>
      )}
    </div>
  )
}

// summariseNode returns one human-readable line per node type for the card.
function summariseNode(n: IVRNode): string {
  const cfg = (n.config ? (Struct.toJson(n.config) as Record<string, unknown>) : {}) as Record<string, unknown>
  switch (n.type) {
    case "play_audio":
      return cfg.audio_path ? `File: ${cfg.audio_path}` : "No audio set"
    case "say":
      return cfg.text ? `“${String(cfg.text).slice(0, 30)}”` : "No text"
    case "menu":
      return cfg.prompt_file ? `Prompt: ${cfg.prompt_file}` : "No prompt"
    case "transfer":
      return cfg.target ? `→ ${cfg.target}` : "No target"
    case "record":
      return cfg.recording_path ? `→ ${cfg.recording_path}` : "No path"
    case "hangup":
      return "Ends the call"
    case "goto":
      return cfg.target_node_id ? `→ ${cfg.target_node_id}` : "No target node"
    default:
      return n.type
  }
}

// ── Inspector ─────────────────────────────────────────────────────────────

interface InspectorProps {
  node: IVRNode
  isStart: boolean
  onMarkStart: () => void
  onChange: (updated: IVRNode) => void
}

function NodeInspector({ node, isStart, onMarkStart, onChange }: InspectorProps) {
  const cfg = nodeConfig(node)
  const setCfg = (key: string, value: unknown) => {
    const next = { ...cfg, [key]: value }
    onChange(setNodeConfig(node, next))
  }
  return (
    <div className="space-y-3">
      <div>
        <div className="text-xs uppercase font-medium text-fg-muted">Node</div>
        <div className="text-sm font-mono">{node.id}</div>
      </div>
      <div>
        <div className="text-xs uppercase font-medium text-fg-muted pb-1">Type</div>
        <div className="text-sm">{node.type}</div>
      </div>
      {!isStart && (
        <button onClick={onMarkStart} className="text-xs underline">
          Mark as start node
        </button>
      )}
      <hr className="border-default" />
      <NodeConfigEditor type={node.type} cfg={cfg} setCfg={setCfg} />
      {node.type === "menu" && (
        <BranchEditor
          branches={node.branches ?? {}}
          onChange={(branches) => onChange({ ...node, branches })}
        />
      )}
    </div>
  )
}

// NodeConfigEditor renders type-specific config inputs. Keeping the schema
// here aligns with the executor's expected keys (see internal/telephony/ivr).
function NodeConfigEditor({
  type,
  cfg,
  setCfg,
}: {
  type: string
  cfg: Record<string, unknown>
  setCfg: (key: string, value: unknown) => void
}) {
  switch (type) {
    case "play_audio":
      return (
        <LabeledInput label="Audio path" value={String(cfg.audio_path ?? "")} onChange={(v) => setCfg("audio_path", v)} />
      )
    case "say":
      return (
        <>
          <LabeledInput label="Text" value={String(cfg.text ?? "")} onChange={(v) => setCfg("text", v)} />
          <LabeledInput label="Voice" value={String(cfg.voice ?? "")} onChange={(v) => setCfg("voice", v)} />
        </>
      )
    case "menu":
      return (
        <>
          <LabeledInput label="Prompt file" value={String(cfg.prompt_file ?? "")} onChange={(v) => setCfg("prompt_file", v)} />
          <LabeledInput label="Timeout (ms)" type="number" value={String(cfg.timeout_ms ?? 5000)} onChange={(v) => setCfg("timeout_ms", Number(v))} />
          <LabeledInput label="Regex" value={String(cfg.regex ?? "")} onChange={(v) => setCfg("regex", v)} />
        </>
      )
    case "transfer":
      return (
        <LabeledInput label="Target" value={String(cfg.target ?? "")} onChange={(v) => setCfg("target", v)} />
      )
    case "record":
      return (
        <>
          <LabeledInput label="Recording path" value={String(cfg.recording_path ?? "")} onChange={(v) => setCfg("recording_path", v)} />
          <LabeledInput label="Max seconds" type="number" value={String(cfg.max_seconds ?? 0)} onChange={(v) => setCfg("max_seconds", Number(v))} />
        </>
      )
    case "goto":
      return (
        <LabeledInput label="Target node id" value={String(cfg.target_node_id ?? "")} onChange={(v) => setCfg("target_node_id", v)} />
      )
    case "hangup":
    default:
      return <div className="text-xs text-fg-muted">No config required.</div>
  }
}

function LabeledInput({
  label,
  value,
  onChange,
  type = "text",
}: {
  label: string
  value: string
  onChange: (v: string) => void
  type?: string
}) {
  return (
    <label className="block space-y-1">
      <span className="text-xs uppercase font-medium text-fg-muted">{label}</span>
      <Input type={type} value={value} onChange={(e) => onChange(e.target.value)} className="text-sm" />
    </label>
  )
}

// BranchEditor lets the user add named branches (typically digits) to a menu
// node. Each branch maps a digit to a target node ID via the canvas edges, so
// adding a branch here exposes a new source handle on the menu card.
function BranchEditor({
  branches,
  onChange,
}: {
  branches: Record<string, string>
  onChange: (next: Record<string, string>) => void
}) {
  const [newKey, setNewKey] = useState("")
  return (
    <div className="space-y-2 pt-2 border-t">
      <div className="text-xs uppercase font-medium text-fg-muted">Branches</div>
      {Object.keys(branches).length === 0 && (
        <div className="text-xs text-fg-muted">Add a digit (e.g. 1) to expose a branch handle.</div>
      )}
      {Object.keys(branches).sort().map((k) => (
        <div key={k} className="flex items-center gap-2 text-sm">
          <span className="font-mono">{k}</span>
          <button
            className="text-xs underline text-fg-muted hover:text-danger"
            onClick={() => {
              const next = { ...branches }
              delete next[k]
              onChange(next)
            }}
          >
            remove
          </button>
        </div>
      ))}
      <div className="flex gap-2">
        <Input
          value={newKey}
          onChange={(e) => setNewKey(e.target.value)}
          placeholder="Add digit"
          className="text-sm"
        />
        <button
          onClick={() => {
            if (!newKey) return
            onChange({ ...branches, [newKey]: "" })
            setNewKey("")
          }}
          className="px-2 py-1 text-xs bg-subtle rounded-xs"
        >
          Add
        </button>
      </div>
    </div>
  )
}

// positionForIndex spreads loaded nodes in a simple grid so the canvas isn't
// stacked at (0, 0). Real layout (dagre, elk) is a follow-up.
function positionForIndex(i: number): { x: number; y: number } {
  const cols = 3
  const col = i % cols
  const row = Math.floor(i / cols)
  return { x: 100 + col * 240, y: 60 + row * 180 }
}
