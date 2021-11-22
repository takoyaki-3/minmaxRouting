package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/takoyaki-3/minmaxRouting"
)

func main(){
	g := makeTestGraph()

	MinMaxRouting(g,Query{
		FromNode: 0,
		ToNode: 6,
		NWeight: 2,
	})
}

type Query struct {
	NWeight int
	FromNode minmaxrouting.NodeIdType
	ToNode minmaxrouting.NodeIdType

}

func MinMaxRouting(g *minmaxrouting.Graph,query Query){

	pos := minmaxrouting.NodeIdType(query.FromNode)
	toNode := minmaxrouting.NodeIdType(query.ToNode)

	// マーティン拡張による最大・最小の
	memo := make([][]CB,len(g.Nodes))

	// 初期化
	memo[pos] = append(memo[pos], CB{
		BeforeNode: -1,
		BeforeIndex: -1,
		Weight: *minmaxrouting.NewWeight(query.NWeight),
	})

	que := Que{}
	que.Add(query.FromNode,*minmaxrouting.NewWeight(query.NWeight),0)
	for que.Len() > 0 {
		pos,i := que.Get()
		if pos == toNode {
			continue
		}

		for _,edgeId := range g.Nodes[pos].FromEdgeIds{
			edge := g.Edges[edgeId]

			newW := WeightAdder(memo[pos][i].Weight,edge.Weight)
			flag := true
			for index,m := range memo[edge.ToId]{
				cmp := CompWeight(m.Weight,newW)
				if cmp < 0{
					flag = false
					break
				} else if cmp > 0 {
					// 悪いので排除
					memo[edge.ToId][index].Weight = minmaxrouting.Weight{}
					for i:=0;i<query.NWeight;i++{
						memo[edge.ToId][index].Weight.Weights = append(memo[edge.ToId][index].Weight.Weights, minmaxrouting.MinMax{
							Min: math.MaxInt32/2,
							Max: math.MaxInt32/2,
						})
					}
				}
			}
		
			if flag{
				que.Add(edge.ToId,newW,len(memo[edge.ToId]))
				memo[edge.ToId] = append(memo[edge.ToId], CB{
					BeforeNode: pos,
					BeforeIndex: i,
					Weight: newW,
				})
			}
		}
	}

	for k,v:=range memo{
		fmt.Println(k,len(v))
	}
	fmt.Println("---route---")
	for i,_:=range memo[toNode] {
		pos = toNode
		for pos != -1{
			fmt.Print(pos,"-")
			bc := memo[pos][i]
			pos = bc.BeforeNode
			i = bc.BeforeIndex
		}
		fmt.Println("")
	}
}

func WeightAdder(w1 minmaxrouting.Weight,w2 minmaxrouting.Weight)(w minmaxrouting.Weight){
	for i,_ := range w1.Weights{
		w.Weights = append(w.Weights, minmaxrouting.MinMax{
			Min: w1.Weights[i].Min+w2.Weights[i].Min,
			Max: w1.Weights[i].Max+w2.Weights[i].Max,
		})
	}
	return w
}

type Que struct {
	n int
	keys []minmaxrouting.NodeIdType
	weights []minmaxrouting.Weight
	indexes []int
}

func (q *Que)Add(key minmaxrouting.NodeIdType,w minmaxrouting.Weight,index int){
	q.weights = append(q.weights, w)
	q.keys = append(q.keys, key)
	q.indexes = append(q.indexes, index)
	q.n++
}

func (q *Que)Len()int{
	return q.n
}

func (q *Que)Get()(minmaxrouting.NodeIdType,int){
	q.n--
	minI := 0
	for i:=1;i<len(q.weights);i++{
		if CompWeight(q.weights[minI],q.weights[i]) > 0{
			minI = i
		}
	}
	rnode := q.keys[minI]
	rindex := q.indexes[minI]
	q.weights[minI] = q.weights[0]
	q.keys[minI] = q.keys[0]
	q.indexes[minI] = q.indexes[0]
	q.weights = q.weights[1:]
	q.keys = q.keys[1:]
	q.indexes = q.indexes[1:]
	return minmaxrouting.NodeIdType(rnode),rindex
}

// w1が良い場合-1、w2が良い場合1、一概に決まらない場合0
func CompWeight(w1 minmaxrouting.Weight,w2 minmaxrouting.Weight)int{
	is1Better := true
	is2Better := true
	
	for i,_:=range w1.Weights{
		if !(w1.Weights[i].Max <= w2.Weights[i].Min) {
			// w1が一概に優っているとは言えない
			is1Better = false
		}
		if !(w2.Weights[i].Max <= w1.Weights[i].Min) {
			// w1が一概に優っているとは言えない
			is2Better = false
		}
	}
	if is1Better == is2Better {
		return 0
	}
	if is1Better {
		return -1
	}
	return 1
}

type CB struct {
	BeforeNode minmaxrouting.NodeIdType
	BeforeIndex int
	Weight minmaxrouting.Weight
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