package routing

import (
	// "fmt"

	"fmt"

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

type Memo [][]int

func MinMaxRouting(g *minmaxrouting.Graph,query Query)(routes []Route,memo Memo,memoList []CB){

	toNode := minmaxrouting.NodeIdType(query.ToNode)

	// マーティン拡張による最大・最小の
	memo = make(Memo,len(g.Nodes))

	// 初期化
	initEdge := minmaxrouting.EdgeIdType(-1)
	memoList = []CB{
		CB{},
		CB{
			BeforeEdgeId: initEdge,
			Weight: *minmaxrouting.NewWeight(query.NWeight),
			IsUse: true,
			Node: query.FromNode,
			BeforeCB: -1,
		},
	}
	memo[query.FromNode] = append(memo[query.FromNode], 1)

	cou := 0
	que := Que{}
	que.Add(1)

	for que.Len() > 0 {
		posCBIndex := que.Get()
		posCb := memoList[posCBIndex]
		// fmt.Println(cou,posCBIndex)
		cou++
		if !posCb.IsUse {
			continue
		}
		if toNode != -1 {
			// fmt.Println("posCb",posCb)
			if posCb.Node == toNode {
				// fmt.Println(pos,memoPosIndex,memoPos.Weight.Weights,que.Len())
				// fmt.Println("Goal!!")
				memo[toNode] = append(memo[toNode], posCBIndex)
				continue
			}
			// 既知のゴールへの重みより大きいか検証
			if !BetterIndex(posCb.Weight,&memo[toNode],&memoList) {
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
			if query.IsSerialNG && posCb.BeforeEdgeId != initEdge && edge.EdgeTypeId == g.Edges[posCb.BeforeEdgeId].EdgeTypeId{
				continue
			}

			// 同一路線を排除
			if !query.IsSerialNG {
				if posCb.BeforeEdgeId != initEdge{
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
			if toNode != -1 && !BetterIndex(newW,&memo[toNode],&memoList) {
				continue
			}
			f := true
			for index,mi := range memo[edge.ToId]{
				if mi < 0{
					continue
				}
				m := memoList[mi]
				if mi >= 0 && memoList[memo[edge.ToId][index]].IsUse {
					cw := CompWeight(m.Weight,newW)
					if cw == -1 {
						f = false
						break
					} else if cw == 1 {
						memoList[memo[edge.ToId][index]].IsUse = false
						memo[edge.ToId][index] = -2
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
			cb := &posCb
			eid := edgeId
			transfer := 0
			for posCBIndex >= 0 && cb.IsUse {
				// if cb == nil {
				// 	flag = false
				// 	break
				// }
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
				if cb.BeforeCB < 0 {
					break
				}
				cb = &memoList[cb.BeforeCB]
				eid = cb.BeforeEdgeId
			}
			
			if !flag {
				continue
			}

			cbIndex := len(memoList)
			memoList = append(memoList, CB{
				Node: edge.ToId,
				BeforeEdgeId: edgeId,
				Weight: newW,
				IsUse: true,
				BeforeCB: posCBIndex,
			})
			cb = &memoList[cbIndex]
			// nilに戻れるかチェック
			// pp := cb
			// p := cb
			// cout := 0
			// for p.BeforeCB >= 0{
			// 	pp = p
			// 	p = &memoList[p.BeforeCB]
			// 	// fmt.Println("cout:",cout)
			// 	cout++
			// }
			// if pp.BeforeEdgeId != -1 {
			// 	fmt.Println("???")
			// }
			que.Add(cbIndex)
			memo[edge.ToId] = append(memo[edge.ToId], cbIndex)
		}
	}

	if toNode != -1 {
		for i,_:=range memo[toNode] {
			if memo[toNode][i] >= 0{
				if BetterIndex(memoList[memo[toNode][i]].Weight,&memo[toNode],&memoList) {
					routes = append(routes, *GetRoute(memo[toNode][i],&memoList))
					fmt.Println("add")
				}
			}
		}
	}

	return routes,memo,memoList
}

func GetRoutes(cbs *[]*CB,memoList *[]CB)[]Route{
	routes := []Route{}
	for _,cb := range *cbs{
		if Better(cb.Weight,cbs) {
			r := GetRoute(cb.BeforeCB,memoList)
			if r != nil{
				routes = append(routes, *r)
			}
		}
	}
	return routes
}

func GetRouteTree(memo Memo,memoList []CB)(*pb.RouteTree){
	tree := &pb.RouteTree{}
	tree.Leaves = make([]*pb.Leaf, len(memoList))
	for index,v:=range memoList{
		wight := &pb.Weight{}
		for _,w := range v.Weight.Weights {
			wight.Max = append(wight.Max, int32(w.Max))
			wight.Min = append(wight.Min, int32(w.Min))
		}
		tree.Leaves[index] = CB2Leaf(&v,wight)
	}
	return tree
}

func CB2Leaf(cb *CB,wight *pb.Weight)*pb.Leaf{
	if cb.BeforeCB == 0{
		cb.BeforeCB = -3
	}
	return &pb.Leaf{
		IsUse: cb.IsUse,
		NodeId: int32(cb.Node),
		Index: int32(cb.BeforeCB),
		BeforeEdgeId: int32(cb.BeforeEdgeId),
		Weight: wight,
	}
}

func GetRoute(pos int,memoList *[]CB)*Route{
	route := Route{
		Weight: (*memoList)[pos].Weight,
	}
	for {
		if !(*memoList)[pos].IsUse {
			return nil
		}
		route.Nodes = append([]minmaxrouting.NodeIdType{(*memoList)[pos].Node},route.Nodes...)
		if (*memoList)[pos].BeforeCB < 0 {
			break
		}
		pos = (*memoList)[pos].BeforeCB
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
func BetterIndex(weight minmaxrouting.Weight,ws *[]int,memoList *[]CB)bool{
	for _,w := range *ws{
		if w >= len(*memoList) {
			fmt.Println("range over",len(*memoList),w)
			continue
		}
		if w >= 0 && len((*memoList)[w].Weight.Weights) > 0{
			if CompWeight(weight,(*memoList)[w].Weight) == 1 {
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
	Cb []int
}

func (q *Que)Add(cb int){
	q.Cb = append(q.Cb, cb)
}

func (q *Que)Len()int{
	return len(q.Cb)
}

func (q *Que)Get()(int){
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
	BeforeCB int
}

