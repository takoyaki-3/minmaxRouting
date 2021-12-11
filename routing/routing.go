package routing

import (
	"github.com/takoyaki-3/minmaxRouting"
	pb "github.com/takoyaki-3/minmaxRouting/pb"
)

type Query struct {
	NWeight int
	FromNode minmaxrouting.NodeIdType
	ToNode minmaxrouting.NodeIdType
	MaxTransfer int
	IsSerialNG bool
}

type Route struct {
	Nodes []minmaxrouting.NodeIdType
	Weight minmaxrouting.Weight
	SubNodes [][]minmaxrouting.NodeIdType
}

type Memo [][]*CB

func MinMaxRouting(g *minmaxrouting.Graph,query Query)(routes []Route,memo Memo){

	toNode := minmaxrouting.NodeIdType(query.ToNode)

	// マーティン拡張による最大・最小の
	memo = make(Memo,len(g.Nodes))

	// 初期化
	initCb := &CB{
		BeforeEdgeId: -1,
		Weight: *minmaxrouting.NewWeight(query.NWeight),
		IsUse: true,
		Node: query.FromNode,
	}
	memo[query.FromNode] = append(memo[query.FromNode], initCb)

	cou := 0
	que := Que{}
	que.Add(initCb)

	for que.Len() > 0 {
		posCb := que.Get()
		if !posCb.IsUse {
			continue
		}
		if toNode != -1 {
			if posCb.Node == toNode {
				// fmt.Println(pos,memoPosIndex,memoPos.Weight.Weights,que.Len())
				memo[toNode] = append(memo[toNode], posCb)
				continue
			}
			// 既知のゴールへの重みより大きいか検証
			if !Better(posCb.Weight,&memo[toNode]) {
				posCb.IsUse = false
				continue
			}
		}

		for _,edgeId := range g.Nodes[posCb.Node].FromEdgeIds{
			edge := g.Edges[edgeId]
			if edge.ToId == posCb.Node{
				// 移動しない辺の為
				continue
			}
			if query.IsSerialNG && posCb.BeforeEdgeId != -1 && edge.EdgeTypeId == g.Edges[posCb.BeforeEdgeId].EdgeTypeId{
				continue
			}

			// 同一路線を排除
			if !query.IsSerialNG {
				if posCb.BeforeEdgeId != -1{
					beforeUseTrips := g.Edges[posCb.BeforeEdgeId].UseTrips
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
			}
			
			newW := WeightAdder(posCb.Weight,edge.Weight)
			// 既知のゴールへの重みより大きいか検証
			if toNode != -1 && !Better(newW,&memo[toNode]) {
				continue
			}
			f := true
			for index,m := range memo[edge.ToId]{
				if m != nil {
					cw := CompWeight(m.Weight,newW)
					if cw == -1 {
						f = false
						break
					} else if cw == 1 {
						memo[edge.ToId][index].IsUse = false
						memo[edge.ToId][index] = nil
					}
				}
			}
			if !f{
				continue
			}

			flag := true
			// 辺の経由駅が訪問済みの場合
			for _,n := range edge.ViaNodes{
				if edge.ToId == n{
					flag = false
					break
				}
			}
			if !flag{
				continue
			}
			// 既に訪問済みか検証
			cb := posCb
			eid := edgeId
			transfer := 0
			for cb != nil && cb.IsUse {
				if cb == nil {
					flag = false
					break
				}				
				// 乗換回数の上限検査
				transfer++
				if toNode != -1{
					if edge.ToId != toNode && transfer > query.MaxTransfer - 1 {
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
				if cb.Node == edge.ToId{
					flag = false
					break
				}

				if eid != -1{
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
				}
				cb = cb.BeforeCB
				if cb == nil {
					break
				}
				eid = cb.BeforeEdgeId
			}
			
			if !flag {
				continue
			}

			cou++
			cb = &CB{
				Node: edge.ToId,
				BeforeEdgeId: edgeId,
				Weight: newW,
				IsUse: true,
				BeforeCB: posCb,
			}
			que.Add(cb)
			memo[edge.ToId] = append(memo[edge.ToId], cb)
		}
	}

	if toNode != -1 {
		for i,_:=range memo[toNode] {
			if memo[toNode][i] != nil{
				if Better(memo[toNode][i].Weight,&memo[toNode]) {
					routes = append(routes, *GetRoute(memo[toNode][i]))
				}	
			}
		}
	}

	return routes,memo
}

func GetRoutes(cbs *[]*CB)[]Route{
	routes := []Route{}
	for _,cb := range *cbs{
		if Better(cb.Weight,cbs) {
			r := GetRoute(cb.BeforeCB)
			if r != nil{
				routes = append(routes, *r)
			}
		}
	}
	return routes
}

func GetRouteTree(memo Memo)(*pb.RouteTree){
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
				BeforeEdgeId: int32(v.BeforeEdgeId),
				Weight: wight,
			})
		}
	}
	return tree
}

func GetRoute(cb *CB)*Route{
	route := Route{
		Weight: cb.Weight,
	}
	for cb != nil {
		if !cb.IsUse {
			return nil
		}
		route.Nodes = append([]minmaxrouting.NodeIdType{cb.Node},route.Nodes...)
		cb = cb.BeforeCB
	}
	return &route
}

func Better(weight minmaxrouting.Weight,ws *[]*CB)bool{
	for _,w := range *ws{
		if w != nil{
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
	Cb []*CB
}

func (q *Que)Add(cb *CB){
	q.Cb = append(q.Cb, cb)
}

func (q *Que)Len()int{
	return len(q.Cb)
}

func (q *Que)Get()(*CB){
	cb := q.Cb[0]
	q.Cb = q.Cb[1:]
	return cb
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
	Node minmaxrouting.NodeIdType
	BeforeEdgeId minmaxrouting.EdgeIdType
	Weight minmaxrouting.Weight
	IsUse bool
	BeforeCB *CB
}

