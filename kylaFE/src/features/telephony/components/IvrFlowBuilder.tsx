import { useCallback, useMemo, useState } from "react"
import {
  ReactFlow,
  Background,
  Controls,
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
import { Link } from "react-router-dom"
import {
  IconArrowLeft,
  IconDeviceFloppy,
  IconLoader2,
  IconTree,
} from "@tabler/icons-react"
import { toast } from "sonner"
import {
  CardSkeleton,
  ErrorState,
} from "@/design-system"
import { cn } from "@/lib/utils"
import { Input } from "@/components/ui/input"
import { useIvrFlow } from "../hooks/queries"
import { updateIvrFlow } from "../api/ivr"
import { IvrNodeType, type IvrFlowNode } from "@/pb/call_ivr_flow"

/**
 * IvrFlowBuilder — React Flow canvas for ordering IVR menus + triggers.
 *
 * Each IvrFlowNode wraps a reference (`nodeId`) to either a menu (CRUD
 * elsewhere) or a trigger. F7 ships the visual ordering + saving of the
 * graph; the menu/trigger editors are linked to from each node and
 * land alongside the queue/extension editors in F7.x.
 */
export function IvrFlowBuilder({ id }: { id: string }) {
  const flow = useIvrFlow(id)

  if (flow.isPending) {
    return (
      <div className="p-6">
        <CardSkeleton lines={4} />
      </div>
    )
  }
  if (flow.isError || !flow.data) {
    return (
      <div className="p-12">
        <ErrorState
          title="Couldn't load IVR flow"
          description={(flow.error as Error | undefined)?.message ?? "Not found"}
          onRetry={() => void flow.refetch()}
        />
      </div>
    )
  }

  return <Editor key={flow.data.id} initial={flow.data} />
}

interface NodeData {
  label: string
  nodeType: IvrNodeType
  nodeId: string
}

function Editor({
  initial,
}: {
  initial: { id: string; name: string; description: string; nodes: IvrFlowNode[] }
}) {
  const [name, setName] = useState(initial.name)
  const [description, setDescription] = useState(initial.description)
  const [dirty, setDirty] = useState(false)
  const [saving, setSaving] = useState(false)

  const [nodes, setNodes] = useState<Node[]>(() =>
    initial.nodes.map((n, i) => ({
      id: n.id || `node-${i}`,
      type: "default",
      position: { x: 240 + (i % 3) * 220, y: 60 + Math.floor(i / 3) * 140 },
      data: {
        label: n.nodeType === IvrNodeType.IVR_NODE_MENU ? "Menu" : "Trigger",
        nodeType: n.nodeType,
        nodeId: n.nodeId,
      } satisfies NodeData,
    })),
  )
  const [edges, setEdges] = useState<Edge[]>(() => {
    const xs: Edge[] = []
    for (let i = 0; i < (initial.nodes.length ?? 0) - 1; i++) {
      xs.push({
        id: `e-${i}`,
        source: initial.nodes[i]!.id || `node-${i}`,
        target: initial.nodes[i + 1]!.id || `node-${i + 1}`,
        type: "smoothstep",
      })
    }
    return xs
  })

  const onNodesChange = useCallback(
    (changes: NodeChange[]) => {
      setDirty(true)
      setNodes((n) => applyNodeChanges(changes, n))
    },
    [],
  )
  const onEdgesChange = useCallback(
    (changes: EdgeChange[]) => {
      setDirty(true)
      setEdges((e) => applyEdgeChanges(changes, e))
    },
    [],
  )
  const onConnect = useCallback(
    (c: Connection) => {
      setDirty(true)
      setEdges((e) => addEdge({ ...c, type: "smoothstep" }, e))
    },
    [],
  )

  const addNode = (type: IvrNodeType) => {
    const id = `node-${Date.now()}`
    setNodes((ns) => [
      ...ns,
      {
        id,
        type: "default",
        position: { x: 240, y: 60 + ns.length * 30 },
        data: {
          label: type === IvrNodeType.IVR_NODE_MENU ? "Menu" : "Trigger",
          nodeType: type,
          nodeId: "",
        } satisfies NodeData,
      },
    ])
    setDirty(true)
  }

  const onSave = async () => {
    setSaving(true)
    try {
      // Reconstruct the ordered list via topological walk of the
      // current edges. Disconnected nodes append at the end so the
      // user doesn't lose work.
      const order = topoOrder(nodes, edges)
      const payload: IvrFlowNode[] = order.map((n) => {
        const d = n.data as unknown as NodeData
        return {
          id: n.id,
          nodeId: d.nodeId,
          nodeType: d.nodeType,
          nodePosition: { x: n.position.x, y: n.position.y } as never,
        } as unknown as IvrFlowNode
      })
      await updateIvrFlow({
        id: initial.id,
        name,
        description,
        nodes: payload,
      })
      setDirty(false)
      toast.success("Flow saved")
    } catch (err) {
      toast.error((err as Error).message)
    } finally {
      setSaving(false)
    }
  }

  const headerActions = useMemo(
    () => (
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={() => addNode(IvrNodeType.IVR_NODE_MENU)}
          className="inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md border border-border hover:bg-subtle"
        >
          + Menu
        </button>
        <button
          type="button"
          onClick={() => addNode(IvrNodeType.IVR_NODE_TRIGGER)}
          className="inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md border border-border hover:bg-subtle"
        >
          + Trigger
        </button>
        <button
          type="button"
          onClick={() => void onSave()}
          disabled={!dirty || saving}
          className={cn(
            "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
            "bg-accent text-accent-fg hover:bg-accent-hover",
            "disabled:opacity-40 disabled:pointer-events-none",
          )}
        >
          {saving ? (
            <IconLoader2 className="size-3.5 animate-spin" />
          ) : (
            <IconDeviceFloppy className="size-3.5" />
          )}
          {dirty ? "Save" : "Saved"}
        </button>
      </div>
    ),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [dirty, saving, name, description, nodes, edges],
  )

  return (
    <div className="h-full flex flex-col">
      <header className="flex items-center gap-3 px-4 py-2 border-b border-border bg-canvas">
        <Link
          to="/ivr"
          aria-label="Back"
          className="inline-flex items-center justify-center size-7 rounded-sm text-fg-secondary hover:text-fg hover:bg-subtle"
        >
          <IconArrowLeft className="size-4" />
        </Link>
        <div className="size-8 rounded-md bg-accent-subtle text-fg-secondary flex items-center justify-center shrink-0" aria-hidden>
          <IconTree className="size-4" />
        </div>
        <div className="min-w-0 flex-1 space-y-1">
          <Input
            value={name}
            onChange={(e) => {
              setName(e.target.value)
              setDirty(true)
            }}
            placeholder="Flow name"
            className="h-7 text-lg font-semibold border-0 bg-transparent shadow-none px-0"
          />
          <Input
            value={description}
            onChange={(e) => {
              setDescription(e.target.value)
              setDirty(true)
            }}
            placeholder="Description"
            className="h-6 text-base border-0 bg-transparent shadow-none px-0 text-fg-muted"
          />
        </div>
        {headerActions}
      </header>

      <div className="flex-1 min-h-0">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          proOptions={{ hideAttribution: true }}
          fitView
          fitViewOptions={{ padding: 0.2 }}
          defaultEdgeOptions={{
            type: "smoothstep",
            style: { stroke: "var(--border-strong)", strokeWidth: 1.5 },
          }}
        >
          <Background gap={16} size={1} color="var(--border-default)" />
          <Controls
            position="bottom-right"
            showInteractive={false}
            className="!bg-surface !border-border"
          />
        </ReactFlow>
      </div>
    </div>
  )
}

/** Topological order driven by edges; disconnected nodes appended. */
function topoOrder(nodes: Node[], edges: Edge[]): Node[] {
  const indeg = new Map<string, number>()
  const adj = new Map<string, string[]>()
  for (const n of nodes) {
    indeg.set(n.id, 0)
    adj.set(n.id, [])
  }
  for (const e of edges) {
    indeg.set(e.target, (indeg.get(e.target) ?? 0) + 1)
    adj.get(e.source)?.push(e.target)
  }
  const queue: string[] = []
  for (const [id, d] of indeg) if (d === 0) queue.push(id)
  const out: Node[] = []
  const byId = new Map(nodes.map((n) => [n.id, n]))
  while (queue.length) {
    const id = queue.shift()!
    const n = byId.get(id)
    if (n) out.push(n)
    for (const next of adj.get(id) ?? []) {
      indeg.set(next, (indeg.get(next) ?? 1) - 1)
      if (indeg.get(next) === 0) queue.push(next)
    }
  }
  // Append any cycle-stranded nodes so the user doesn't lose work.
  for (const n of nodes) {
    if (!out.includes(n)) out.push(n)
  }
  return out
}
