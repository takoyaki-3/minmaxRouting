package routing

import (
	"fmt"

	"github.com/takoyaki-3/minmaxRouting"
	pb "github.com/takoyaki-3/minmaxRouting/pb"
)

type Query struct {
	NWeight int
	FromNode minmaxrouting.NodeIdType
	ToNode minmaxrouting.NodeIdType
	MaxTransfer int
}

type Route struct {
	Nodes []minmaxrouting.NodeIdType
	Weight minmaxrouting.Weight
	SubNodes [][]minmaxrouting.NodeIdType
}

func MinMaxRouting(g *minmaxrouting.Graph,query Query)(routes []Route,memo [][]CB){

	pos := minmaxrouting.NodeIdType(query.FromNode)
	toNode := minmaxrouting.NodeIdType(query.ToNode)

	// マーティン拡張による最大・最小の
	memo = make([][]CB,len(g.Nodes))

	// 初期化
	memo[pos] = append(memo[pos], CB{
		BeforeNode: -1,
		BeforeIndex: -1,
		BeforeEdgeId: -1,
		Weight: *minmaxrouting.NewWeight(query.NWeight),
		IsUse: true,
	})

	cou := 0
	que := Que{}
	que.Add(query.FromNode,*minmaxrouting.NewWeight(query.NWeight),0)
	for que.Len() > 0 {
		pos,memoPosIndex := que.Get()
		memoPos := &memo[pos][memoPosIndex]
		if !memoPos.IsUse{
			continue
		}
		if toNode != -1 {
			if pos == toNode {
				// fmt.Println(pos,i,memo[pos][i].Weight.Weights,que.Len())
				continue
			}
			// 既知のゴールへの重みより大きいか検証
			if !Better(memoPos.Weight,&memo[toNode]) {
				*memoPos = CB{
					IsUse: false,
				}
				continue
			}
		}

		for _,edgeId := range g.Nodes[pos].FromEdgeIds{
			edge := g.Edges[edgeId]
			if edge.ToId == pos{
				continue
			}

			// 同一路線を排除
			if memoPos.BeforeEdgeId != -1{
				beforeUseTrips := g.Edges[memoPos.BeforeEdgeId].UseTrips
				if len(beforeUseTrips) > 0{
					befS := -1
					for tripIndex,trip := range edge.UseTrips{
						if trip == beforeUseTrips[0]{
							befS = tripIndex
						}
					}
					if befS != -1{
						flag := false
						for index:=0;index<len(edge.UseTrips)-befS;index++{
							if len(beforeUseTrips) == index{
								flag = true
								break
							}
							if edge.UseTrips[befS+index] != beforeUseTrips[index]{
								flag = true
								break
							}
						}
						if !flag {
							continue
						}
					}
				}
			}

			newW := WeightAdder(memoPos.Weight,edge.Weight)
			// 既知のゴールへの重みより大きいか検証
			if toNode != -1 && !Better(newW,&memo[toNode]) {
				continue
			}
			f := true
			for index,m := range memo[edge.ToId]{
				if m.IsUse {
					cw := CompWeight(m.Weight,newW)
					if cw == -1 {
						f = false
						break
					} else if cw == 1 {
						memo[edge.ToId][index] = CB{
							IsUse: false,
						}
					}
				}
			}
			if !f{
				continue
			}

			flag := true
			// // 辺の経由駅が訪問済みの場合
			// for _,n := range edge.ViaNodes{
			// 	if edge.ToId == n{
			// 		flag = false
			// 		break
			// 	}
			// }
			// if !flag{
			// 	continue
			// }
			// 既に訪問済みか検証
			p := pos
			ind := memoPosIndex
			eid := edgeId
			transfer := 0
			for p != -1{
				if !memo[p][ind].IsUse{
					flag = false
					break
				}
				// 乗換回数の上限検査
				transfer++
				if toNode != -1{
					if transfer > query.MaxTransfer - 1 {
						flag = false
						break
					}
				} else {
					if transfer > query.MaxTransfer {
						flag = false
						break
					}
				}
				// 既に訪問済みの場合
				if p == edge.ToId{
					flag = false
					break
				}

				if eid != edgeId{
					for _,vnode := range g.Edges[eid].ViaNodes{
						if vnode == edge.ToId{
							flag = false
							break
						}
						for _,n := range edge.ViaNodes{
							if vnode == n {
								flag = false
								break
							}
						}
						if !flag {
							break
						}	
					}
					if !flag {
						break
					}
				}
				m := &memo[p][ind]
				p = m.BeforeNode
				ind = m.BeforeIndex
				eid = m.BeforeEdgeId
			}
			if !flag {
				continue
			}
			if false && cou % 1000 == 0{
				fmt.Println("add",cou,newW)
			}
			cou++
			que.Add(edge.ToId,newW,len(memo[edge.ToId]))
			memo[edge.ToId] = append(memo[edge.ToId], CB{
				BeforeNode: pos,
				BeforeIndex: memoPosIndex,
				BeforeEdgeId: edgeId,
				Weight: newW,
				IsUse: true,
			})
		}
	}

	if toNode != -1 {
		for i,_:=range memo[toNode] {
			if memo[toNode][i].IsUse{
				if Better(memo[toNode][i].Weight,&memo[toNode]) {
					routes = append(routes, *GetRoute(memo,toNode,i))
				}	
			}
		}
	// } else {
	// 	for toNode,_ := range g.Nodes{
	// 		for i,_:=range memo[toNode] {
	// 			if Better(memo[toNode][i].Weight,&memo[toNode]) {
	// 				routes = append(routes, *GetRoute(memo,minmaxrouting.NodeIdType(toNode),i))
	// 			}
	// 		}
	// 	}
	}

	return routes,memo
}

func GetRoutes(memo [][]CB)[]Route{
	routes := []Route{}
	for toNode,v := range memo{
		for i,_:=range v {
			if Better(v[i].Weight,&v) {
				routes = append(routes, *GetRoute(memo,minmaxrouting.NodeIdType(toNode),i))
			}
		}	
	}
	return routes
}

func GetRouteTree(memo [][]CB)(*pb.RouteTree){
	tree := &pb.RouteTree{}
	for nid,v:=range memo{
		for index,v:=range v{
			wight := &pb.Weight{}
			for _,w := range v.Weight.Weights{
				wight.Max = append(wight.Max, int32(w.Max))
				wight.Min = append(wight.Min, int32(w.Min))
			}
			tree.Leaves = append(tree.Leaves, &pb.Leaf{
				NodeId: int32(nid),
				Index: int32(index),
				BeforeNodeId: int32(v.BeforeNode),
				BeforeIndex: int32(v.BeforeIndex),
				BeforeEdgeId: int32(v.BeforeEdgeId),
				Weight: wight,
			})
		}
	}
	return tree
}

func GetRoute(memo [][]CB,pos minmaxrouting.NodeIdType,i int)*Route{
	route := Route{
		Weight: memo[pos][i].Weight,
	}
	pos = pos
	for pos != -1{
		route.Nodes = append([]minmaxrouting.NodeIdType{pos},route.Nodes...)
		bc := memo[pos][i]
		pos = bc.BeforeNode
		i = bc.BeforeIndex
	}
	return &route
}

func Better(weight minmaxrouting.Weight,ws *[]CB)bool{
	for _,w := range *ws{
		if w.IsUse{
			if CompWeight(weight,w.Weight) == 1 {
				return false
			}	
		}
	}
	return true
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
	rnode := q.keys[minI]
	rindex := q.indexes[minI]
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
	BeforeEdgeId minmaxrouting.EdgeIdType
	Weight minmaxrouting.Weight
	IsUse bool
}

