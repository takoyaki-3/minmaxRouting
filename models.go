package minmaxrouting

type Weight struct {
	Weights []MinMax
}

type SingleCostType int

type MinMax struct {
	Min SingleCostType
	Max SingleCostType
}

func NewWeight(n int)*Weight{
	var w Weight
	w.Weights = make([]MinMax, n)
	return &w
}

type EdgeIdType int
type NodeIdType int

type Edge struct {
	FromId NodeIdType
	ToId NodeIdType
	Weight Weight

	ViaNodes []NodeIdType
}

type Node struct {
	FromEdgeIds []EdgeIdType
	ToEdgeIds []EdgeIdType
}

type Graph struct {
	Nodes []Node
	Edges []Edge
}
