import dagre from "@dagrejs/dagre"
import type { Node, Edge } from "@xyflow/react"

/**
 * Auto-lays out a React Flow graph using dagre's top-to-bottom hierarchical
 * algorithm. Returns a new node array with `position` overwritten; edges are
 * unchanged.
 *
 * Used by the IVR visual builder's "Auto-layout" button so operators don't
 * have to drag every node into place after editing the graph.
 *
 * The defaults (180×80 node box, 60px ranksep, 40px nodesep) are tuned for
 * the IvrCanvasNode component's typical dimensions. Tweak there if you
 * change the card's footprint.
 */
export function autoLayoutGraph(nodes: Node[], edges: Edge[]): Node[] {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: "TB", nodesep: 40, ranksep: 60 })

  for (const n of nodes) {
    g.setNode(n.id, { width: 180, height: 80 })
  }
  for (const e of edges) {
    g.setEdge(e.source, e.target)
  }
  dagre.layout(g)

  return nodes.map((n) => {
    const layouted = g.node(n.id)
    if (!layouted) return n
    return {
      ...n,
      position: { x: layouted.x - 90, y: layouted.y - 40 },
    }
  })
}
