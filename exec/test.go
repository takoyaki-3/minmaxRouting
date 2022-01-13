package main

import (
	"fmt"
	"math/rand"
	"github.com/takoyaki-3/minmaxRouting"
	"github.com/takoyaki-3/minmaxRouting/routing"
)

func main(){
	g := makeTestGraph()

	routes,_,_ := routing.MinMaxRouting(g,routing.Query{
		FromNode: 0,
		ToNode: 6,
		NWeight: 2,
		MaxTransfer: 100,
	})

	fmt.Println("---routes---")
	for _,route := range routes{
		for _,n:=range route.Nodes{
			fmt.Print(n," ")
		}
		fmt.Println("")
	}
}

func makeTestGraph()*minmaxrouting.Graph{
	g := &minmaxrouting.Graph{}

	g.Nodes = append(g.Nodes, minmaxrouting.Node{})
	g.Nodes = append(g.Nodes, minmaxrouting.Node{})
	g.Nodes = append(g.Nodes, minmaxrouting.Node{})
	g.Nodes = append(g.Nodes, minmaxrouting.Node{})
	g.Nodes = append(g.Nodes, minmaxrouting.Node{})
	g.Nodes = append(g.Nodes, minmaxrouting.Node{})
	g.Nodes = append(g.Nodes, minmaxrouting.Node{})

	rand.Int()

	const nWeight = 2

	g.Edges = append(g.Edges, minmaxrouting.Edge{
		FromId: 0,
		ToId: 1,
		Weight: RandWeight(nWeight),
	})
	g.Edges = append(g.Edges, minmaxrouting.Edge{
		FromId: 0,
		ToId: 2,
		Weight: RandWeight(nWeight),
	})
	g.Edges = append(g.Edges, minmaxrouting.Edge{
		FromId: 0,
		ToId: 3,
		Weight: RandWeight(nWeight),
	})
	g.Edges = append(g.Edges, minmaxrouting.Edge{
		FromId: 1,
		ToId: 3,
		Weight: RandWeight(nWeight),
	})
	g.Edges = append(g.Edges, minmaxrouting.Edge{
		FromId: 1,
		ToId: 4,
		Weight: RandWeight(nWeight),
	})
	g.Edges = append(g.Edges, minmaxrouting.Edge{
		FromId: 2,
		ToId: 3,
		Weight: RandWeight(nWeight),
	})
	g.Edges = append(g.Edges, minmaxrouting.Edge{
		FromId: 2,
		ToId: 5,
		Weight: RandWeight(nWeight),
	})
	g.Edges = append(g.Edges, minmaxrouting.Edge{
		FromId: 3,
		ToId: 4,
		Weight: RandWeight(nWeight),
	})
	g.Edges = append(g.Edges, minmaxrouting.Edge{
		FromId: 3,
		ToId: 5,
		Weight: RandWeight(nWeight),
	})
	g.Edges = append(g.Edges, minmaxrouting.Edge{
		FromId: 3,
		ToId: 6,
		Weight: RandWeight(nWeight),
	})
	g.Edges = append(g.Edges, minmaxrouting.Edge{
		FromId: 4,
		ToId: 6,
		Weight: RandWeight(nWeight),
	})
	g.Edges = append(g.Edges, minmaxrouting.Edge{
		FromId: 5,
		ToId: 6,
		Weight: RandWeight(nWeight),
	})

	for _,e := range g.Edges {
		fmt.Println(e.FromId,"->",e.ToId,":",e.Weight.Weights)
	}

	elen := len(g.Edges)
	for i:=0;i<elen;i++{
		edge := g.Edges[i]
		g.Edges = append(g.Edges, minmaxrouting.Edge{
			FromId: edge.ToId,
			ToId: edge.FromId,
			Weight: edge.Weight,
		})
	}

	for i,e := range g.Edges {
		g.Nodes[e.FromId].FromEdgeIds = append(g.Nodes[e.FromId].FromEdgeIds, minmaxrouting.EdgeIdType(i))
		g.Nodes[e.ToId].ToEdgeIds = append(g.Nodes[e.ToId].ToEdgeIds, minmaxrouting.EdgeIdType(i))
	}
	return g
}

func RandWeight(n int)(w minmaxrouting.Weight){

	for i:=0;i<n;i++{
		a := rand.Int()%10+1
		b := rand.Int()%5
		w.Weights = append(w.Weights, minmaxrouting.MinMax{
			Min: minmaxrouting.SingleCostType(a),
			Max: minmaxrouting.SingleCostType(a+b),
		})		
	}
	return w
}