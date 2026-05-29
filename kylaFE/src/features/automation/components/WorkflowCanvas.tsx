import { useCallback, useEffect, useMemo, useState } from "react"
import {
  ReactFlow,
  Background,
  Controls,
  type Node,
  type Edge,
  type Connection,
  addEdge,
  applyNodeChanges,
  applyEdgeChanges,
  type NodeChange,
  type EdgeChange,
} from "@xyflow/react"
import "@xyflow/react/dist/style.css"
import { TriggerNode } from "../nodes/TriggerNode"
import { ConditionNode } from "../nodes/ConditionNode"
import { ActionNode } from "../nodes/ActionNode"
import { NodePalette } from "./NodePalette"
import { NodeInspector } from "./NodeInspector"
import type { ActionType, Condition, TriggerType } from "../utils/actions"
import type { ActionNode as ActionModel, WorkflowSpec } from "../utils/types"

/**
 * WorkflowCanvas — visual workflow builder.
 *
 *   ┌──────────┬───────────────────────────────┬──────────┐
 *   │ Palette  │ React Flow canvas              │ Inspector│
 *   └──────────┴───────────────────────────────┴──────────┘
 *
 * The canvas is the source of truth while editing. On save the parent
 * serializes the React Flow graph back into a WorkflowSpec (trigger +
 * conditions + ordered actions) via `extractSpec`.
 */

const nodeTypes = {
  trigger: TriggerNode,
  condition: ConditionNode,
  action: ActionNode,
}

export interface WorkflowCanvasProps {
  initial: WorkflowSpec
  onChange: (next: {
    trigger: { type: TriggerType; config?: Record<string, unknown> } | null
    conditions: Condition[]
    actions: ActionModel[]
  }) => void
}

interface NodeData {
  triggerType?: TriggerType | null
  config?: Record<string, unknown>
  conditions?: Condition[]
  actionType?: ActionType
  params?: Record<string, unknown>
}

export function WorkflowCanvas({ initial, onChange }: WorkflowCanvasProps) {
  const [nodes, setNodes] = useState<Node[]>(() => initialNodes(initial))
  const [edges, setEdges] = useState<Edge[]>(() => initialEdges(initial))
  const [selectedId, setSelectedId] = useState<string | null>(null)

  // Push current graph back to the parent as a typed WorkflowSpec slice.
  useEffect(() => {
    onChange(extractSpec(nodes, edges))
    // We intentionally don't include onChange to avoid re-emitting when
    // the parent re-creates the callback on every render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nodes, edges])

  const onNodesChange = useCallback(
    (changes: NodeChange[]) => setNodes((n) => applyNodeChanges(changes, n)),
    [],
  )
  const onEdgesChange = useCallback(
    (changes: EdgeChange[]) => setEdges((e) => applyEdgeChanges(changes, e)),
    [],
  )
  const onConnect = useCallback(
    (params: Connection) =>
      setEdges((e) => addEdge({ ...params, animated: true }, e)),
    [],
  )

  const selected = useMemo(
    () => nodes.find((n) => n.id === selectedId) ?? null,
    [nodes, selectedId],
  )

  const patchNode = useCallback(
    (id: string, patch: Record<string, unknown>) =>
      setNodes((ns) =>
        ns.map((n) =>
          n.id === id ? { ...n, data: { ...n.data, ...patch } } : n,
        ),
      ),
    [],
  )

  const deleteNode = useCallback((id: string) => {
    setNodes((ns) => ns.filter((n) => n.id !== id))
    setEdges((es) => es.filter((e) => e.source !== id && e.target !== id))
    setSelectedId(null)
  }, [])

  const addTrigger = useCallback(() => {
    const id = `trigger-${Date.now()}`
    setNodes((ns) =>
      ns.filter((n) => n.type !== "trigger").concat({
        id,
        type: "trigger",
        position: { x: 240, y: 40 },
        data: { triggerType: null, config: {} },
      }),
    )
    setSelectedId(id)
  }, [])

  const addCondition = useCallback(() => {
    const id = `cond-${Date.now()}`
    setNodes((ns) => [
      ...ns,
      {
        id,
        type: "condition",
        position: { x: 240, y: 180 + ns.length * 20 },
        data: { conditions: [] },
      },
    ])
    setSelectedId(id)
  }, [])

  const addAction = useCallback((type: ActionType) => {
    const id = `action-${Date.now()}`
    setNodes((ns) => [
      ...ns,
      {
        id,
        type: "action",
        position: { x: 240, y: 320 + ns.length * 60 },
        data: { actionType: type, params: {} },
      },
    ])
    setSelectedId(id)
  }, [])

  return (
    <div className="flex h-full min-h-0 bg-canvas">
      <NodePalette
        onAddTrigger={addTrigger}
        onAddCondition={addCondition}
        onAddAction={addAction}
      />
      <div className="flex-1 min-w-0">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onPaneClick={() => setSelectedId(null)}
          onNodeClick={(_e, node) => setSelectedId(node.id)}
          nodeTypes={nodeTypes}
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
      {selected && (
        <NodeInspector
          node={selected}
          onPatch={patchNode}
          onDelete={deleteNode}
          onClose={() => setSelectedId(null)}
        />
      )}
    </div>
  )
}

// ── Initial graph hydration ──────────────────────────────────────────────────

function initialNodes(spec: WorkflowSpec): Node[] {
  const ns: Node[] = []

  if (spec.trigger) {
    ns.push({
      id: "trigger-1",
      type: "trigger",
      position: { x: 240, y: 40 },
      data: {
        triggerType: spec.trigger.type,
        config: spec.trigger.config ?? {},
      } satisfies NodeData,
    })
  }

  if (spec.conditions.length > 0) {
    ns.push({
      id: "cond-1",
      type: "condition",
      position: { x: 240, y: 180 },
      data: { conditions: spec.conditions } satisfies NodeData,
    })
  }

  spec.actions.forEach((a, i) =>
    ns.push({
      id: a.id || `action-${i}`,
      type: "action",
      position: { x: 240, y: 320 + i * 140 },
      data: { actionType: a.type, params: a.params } satisfies NodeData,
    }),
  )

  return ns
}

function initialEdges(spec: WorkflowSpec): Edge[] {
  const ids: string[] = []
  if (spec.trigger) ids.push("trigger-1")
  if (spec.conditions.length > 0) ids.push("cond-1")
  spec.actions.forEach((a, i) => ids.push(a.id || `action-${i}`))

  const edges: Edge[] = []
  for (let i = 0; i < ids.length - 1; i++) {
    edges.push({
      id: `e-${i}`,
      source: ids[i]!,
      target: ids[i + 1]!,
      type: "smoothstep",
      animated: false,
    })
  }
  return edges
}

// ── Graph → WorkflowSpec extraction ──────────────────────────────────────────

/**
 * Walk the graph from the trigger downward and produce a typed slice
 * of the spec. Disconnected actions are appended at the end so the
 * user doesn't lose work; any unconnected condition is dropped.
 */
function extractSpec(nodes: Node[], edges: Edge[]) {
  const trigger = nodes.find((n) => n.type === "trigger")
  const triggerSpec = trigger
    ? {
        type: (trigger.data as NodeData).triggerType as TriggerType,
        config: (trigger.data as NodeData).config ?? {},
      }
    : null

  // Build adjacency map keyed by source.
  const next = new Map<string, string[]>()
  for (const e of edges) {
    if (!next.has(e.source)) next.set(e.source, [])
    next.get(e.source)!.push(e.target)
  }

  const visited = new Set<string>()
  const conditions: Condition[] = []
  const actions: ActionModel[] = []

  function walk(id: string) {
    if (visited.has(id)) return
    visited.add(id)
    const node = nodes.find((n) => n.id === id)
    if (!node) return
    if (node.type === "condition") {
      const c = (node.data as NodeData).conditions ?? []
      conditions.push(...c)
    }
    if (node.type === "action") {
      actions.push({
        id: node.id,
        type: (node.data as NodeData).actionType as ActionType,
        params: (node.data as NodeData).params ?? {},
      })
    }
    for (const child of next.get(id) ?? []) {
      walk(child)
    }
  }

  if (trigger) walk(trigger.id)

  // Append disconnected actions so users don't lose unsaved nodes.
  for (const n of nodes) {
    if (n.type === "action" && !visited.has(n.id)) {
      actions.push({
        id: n.id,
        type: (n.data as NodeData).actionType as ActionType,
        params: (n.data as NodeData).params ?? {},
      })
    }
  }

  return {
    trigger: triggerSpec && triggerSpec.type
      ? { type: triggerSpec.type, config: triggerSpec.config }
      : null,
    conditions,
    actions,
  }
}
