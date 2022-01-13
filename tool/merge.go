package tool

import (
	minmax "github.com/takoyaki-3/minmaxRouting"
)

type Id2Ids struct {
	Nodeid2nodeids map[minmax.NodeIdType]minmax.NodeIdType
	Edgeid2edgeids map[minmax.EdgeIdType]minmax.EdgeIdType
	Nodeid2componentids map[minmax.NodeIdType]int
	Edgeid2componentids map[minmax.EdgeIdType]int

	OldNodeid2nodeids []map[minmax.NodeIdType]minmax.NodeIdType
	OldEdgeid2edgeids []map[minmax.EdgeIdType]minmax.EdgeIdType
}

func MergeMinMaxGraphs(gs []*minmax.Graph)(*minmax.Graph,*Id2Ids){
	var numNodes,numEdges int

	mg := &minmax.Graph{}

	id2ids := Id2Ids{
		Nodeid2nodeids: map[minmax.NodeIdType]minmax.NodeIdType{},
		Edgeid2edgeids: map[minmax.EdgeIdType]minmax.EdgeIdType{},
		Nodeid2componentids: map[minmax.NodeIdType]int{},
		Edgeid2componentids: map[minmax.EdgeIdType]int{},
		OldNodeid2nodeids: make([]map[minmax.NodeIdType]minmax.NodeIdType, len(gs)),
		OldEdgeid2edgeids: make([]map[minmax.EdgeIdType]minmax.EdgeIdType, len(gs)),
	}

	for j,g:=range gs{
		Nodeid2nodeids := map[minmax.NodeIdType]minmax.NodeIdType{}
		Edgeid2edgeids := map[minmax.EdgeIdType]minmax.EdgeIdType{}
		id2ids.OldNodeid2nodeids[j] = map[minmax.NodeIdType]minmax.NodeIdType{}
		id2ids.OldEdgeid2edgeids[j] = map[minmax.EdgeIdType]minmax.EdgeIdType{}

		for i,_:=range g.Nodes{
			Nodeid2nodeids[minmax.NodeIdType(i)] = minmax.NodeIdType(numNodes+i)
			id2ids.Nodeid2nodeids[minmax.NodeIdType(numNodes+i)] = minmax.NodeIdType(i)
			id2ids.Nodeid2componentids[minmax.NodeIdType(numNodes+i)] = j
			id2ids.OldNodeid2nodeids[j][minmax.NodeIdType(i)] = minmax.NodeIdType(numNodes+i)
		}
		for i,e:=range g.Edges{
			Edgeid2edgeids[minmax.EdgeIdType(i)] = minmax.EdgeIdType(numEdges+i)
			id2ids.Edgeid2edgeids[minmax.EdgeIdType(numEdges+i)] = minmax.EdgeIdType(i)
			id2ids.Edgeid2componentids[minmax.EdgeIdType(numEdges+i)] = j
			id2ids.OldEdgeid2edgeids[j][minmax.EdgeIdType(i)] = minmax.EdgeIdType(numNodes+i)
			mg.Edges = append(mg.Edges, minmax.Edge{
				FromId: Nodeid2nodeids[e.FromId],
				ToId: Nodeid2nodeids[e.ToId],
				Weight: e.Weight,
				ViaNodes: e.ViaNodes,
			})
		}
		numNodes += len(g.Nodes)
		numEdges += len(g.Edges)
	}
	mg.Nodes = make([]minmax.Node, numNodes)
	return mg,&id2ids
}

// func Deduplication(g *minmax.Graph){
// 	rmEdges := []minmax.EdgeIdType{}
// 	for i:=0;i<len(g.Edges);i++{
// 		for j:=i+1;j<len(g.Edges);j++{
// 			if g.Edges[i].FromId == g.
// 		}
// 	}
// }
